package application_test

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel/trace/noop"

	"github.com/cubelitblade/community-v2/backend/services/authz/internal/application"
	"github.com/cubelitblade/community-v2/backend/services/authz/internal/domain"
	"github.com/cubelitblade/community-v2/backend/services/authz/internal/domain/port"
)

var (
	_ port.TupleWriter      = (*FakeTupleWriter)(nil)
	_ port.CacheInvalidator = (*FakeCacheInvalidator)(nil)
)

type FakeTupleWriter struct {
	CalledCount   int
	WrittenTuples []domain.Tuple
	errToReturn   error
}

func (w *FakeTupleWriter) Write(_ context.Context, tuples []domain.Tuple) error {
	w.CalledCount++
	w.WrittenTuples = append(w.WrittenTuples, tuples...)

	return w.errToReturn
}

type FakeCacheInvalidator struct {
	CalledCount     int
	InvalidatedKeys []string
	errToReturn     error
}

func (i *FakeCacheInvalidator) Invalidate(_ context.Context, keys []string) error {
	i.CalledCount++
	i.InvalidatedKeys = append(i.InvalidatedKeys, keys...)

	return i.errToReturn
}

func TestTupleWriter_Write(t *testing.T) {
	t.Parallel()

	tupleAliceEditPost := domain.Tuple{User: UserAlice, Relation: RelationEditPost, Object: PostAlice}
	tupleBobDeletePost := domain.Tuple{User: AdminBob, Relation: RelationDeletePost, Object: PostAlice}
	tupleBobEditPost := domain.Tuple{User: AdminBob, Relation: RelationEditPost, Object: PostBob}

	tests := []struct {
		name string
		// Input.
		tuples []domain.Tuple
		// Optional: force an error on a specific fake.
		writerErr      error
		invalidatorErr error
		// Expected outputs.
		wantErr               error
		wantWriterCalls       int
		wantInvalidatorCalls  int
		wantInvalidatedKeySet []string // nil means skip check
		wantWrittenTupleCount int
	}{
		{
			name:                 "empty tuples, no-op",
			tuples:               nil,
			wantWriterCalls:      0,
			wantInvalidatorCalls: 0,
		},
		{
			name:                  "single tuple, success",
			tuples:                []domain.Tuple{tupleAliceEditPost},
			wantWriterCalls:       1,
			wantInvalidatorCalls:  1,
			wantInvalidatedKeySet: []string{PostAlice, UserAlice},
			wantWrittenTupleCount: 1,
		},
		{
			name: "duplicate keys deduplicated",
			tuples: []domain.Tuple{
				tupleAliceEditPost,
				tupleBobDeletePost,
			},
			wantWriterCalls:       1,
			wantInvalidatorCalls:  1,
			wantInvalidatedKeySet: []string{PostAlice, UserAlice, AdminBob},
			wantWrittenTupleCount: 2,
		},
		{
			name: "totally different tuples",
			tuples: []domain.Tuple{
				tupleAliceEditPost,
				tupleBobEditPost,
			},
			wantWriterCalls:       1,
			wantInvalidatorCalls:  1,
			wantInvalidatedKeySet: []string{PostAlice, PostBob, UserAlice, AdminBob},
			wantWrittenTupleCount: 2,
		},
		{
			name:                  "write succeeds, cache invalidate fails",
			tuples:                []domain.Tuple{tupleAliceEditPost},
			invalidatorErr:        errFake,
			wantErr:               errFake,
			wantInvalidatorCalls:  1,
			wantWriterCalls:       1,
			wantWrittenTupleCount: 1,
		},
		{
			name:                 "writer fails",
			tuples:               []domain.Tuple{tupleAliceEditPost},
			writerErr:            errFake,
			wantErr:              errFake,
			wantInvalidatorCalls: 0,
			wantWriterCalls:      1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			writer := &FakeTupleWriter{errToReturn: tt.writerErr}
			invalidator := &FakeCacheInvalidator{errToReturn: tt.invalidatorErr}

			w := application.NewTupleWriter(
				writer, invalidator, slog.Default(), noop.NewTracerProvider(),
			)
			defer w.Stop(context.Background())

			gotErr := w.Write(context.Background(), tt.tuples)

			if tt.wantErr != nil {
				assert.ErrorIs(t, gotErr, tt.wantErr)
			} else {
				assert.NoError(t, gotErr)
			}

			assert.Equal(t, tt.wantWriterCalls, writer.CalledCount)
			assert.Equal(t, tt.wantInvalidatorCalls, invalidator.CalledCount)

			if tt.wantInvalidatedKeySet != nil {
				assert.ElementsMatch(t, tt.wantInvalidatedKeySet, invalidator.InvalidatedKeys)
			}

			if tt.wantWrittenTupleCount > 0 {
				assert.Len(t, writer.WrittenTuples, tt.wantWrittenTupleCount)
			}
		})
	}
}

func TestTupleWriter_Stop(t *testing.T) {
	t.Parallel()

	tupleAliceEditPost := domain.Tuple{User: UserAlice, Relation: RelationEditPost, Object: PostAlice}

	tests := []struct {
		name                 string
		writeBeforeStop      bool
		writeAfterStop       bool
		waitForAsync         bool
		wantInvalidatorCalls int
	}{
		{
			name:                 "Stop returns promptly",
			writeBeforeStop:      false,
			writeAfterStop:       false,
			waitForAsync:         false,
			wantInvalidatorCalls: 0,
		},
		{
			name:                 "pending async invalidation is dropped after Stop",
			writeBeforeStop:      true,
			writeAfterStop:       false,
			waitForAsync:         true,
			wantInvalidatorCalls: 1,
		},
		{
			name:                 "Write after Stop drops async task but still syncs",
			writeBeforeStop:      false,
			writeAfterStop:       true,
			waitForAsync:         true,
			wantInvalidatorCalls: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			writer := &FakeTupleWriter{}
			invalidator := &FakeCacheInvalidator{}

			w := application.NewTupleWriter(
				writer, invalidator, slog.Default(), noop.NewTracerProvider(),
			)

			if tt.writeBeforeStop {
				err := w.Write(context.Background(), []domain.Tuple{tupleAliceEditPost})
				assert.NoError(t, err)
			}

			stopCtx, stopCancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer stopCancel()

			err := w.Stop(stopCtx)
			assert.NoError(t, err, "Stop should return promptly without timeout error")

			if tt.writeAfterStop {
				err := w.Write(context.Background(), []domain.Tuple{tupleAliceEditPost})
				assert.NoError(t, err)
			}

			if tt.waitForAsync {
				time.Sleep(1 * time.Second)
			}

			assert.Equal(t, tt.wantInvalidatorCalls, invalidator.CalledCount)
		})
	}
}

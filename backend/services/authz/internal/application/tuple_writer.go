package application

import (
	"context"
	"fmt"
	"log/slog"
	"slices"
	"sync"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"github.com/cubelitblade/community-v2/backend/services/authz/internal/domain"
	"github.com/cubelitblade/community-v2/backend/services/authz/internal/domain/port"
)

const (
	cacheInvalidationDelay   = 500 * time.Millisecond
	cacheInvalidationTimeout = 3 * time.Second
	taskQueueSize            = 1024
	workerCount              = 5
)

type invalidationTask struct {
	link trace.Link
	keys []string
}

// TupleWriter coordinates persisting authorization tuples to storage and managing cache consistency.
type TupleWriter struct {
	writer port.TupleWriter
	cache  port.CacheInvalidator
	logger *slog.Logger
	tracer trace.Tracer

	ch      chan *invalidationTask
	stop    chan struct{}
	wg      sync.WaitGroup
	stopped bool
	mu      sync.RWMutex
}

// NewTupleWriter creates a new TupleWriter instance.
func NewTupleWriter(
	writer port.TupleWriter, cacheInvalidator port.CacheInvalidator,
	logger *slog.Logger, traceProvider trace.TracerProvider,
) *TupleWriter {
	tupleWriter := &TupleWriter{
		writer: writer,
		cache:  cacheInvalidator,
		logger: logger.With(slog.String("component", "tuple_writer")),
		tracer: traceProvider.Tracer("authz.application.tuple_writer"),

		ch:      make(chan *invalidationTask, taskQueueSize),
		stop:    make(chan struct{}),
		wg:      sync.WaitGroup{},
		stopped: false,
		mu:      sync.RWMutex{},
	}

	for range workerCount {
		tupleWriter.wg.Add(1)
		go tupleWriter.runDelayedInvalidation()
	}

	return tupleWriter
}

// Write persists authorization tuples and ensures cache consistency.
//
// The side-cache handles the majority of consistency issues. The remaining
// rare edge cases—caused by concurrent slow reads—are addressed by a
// best-effort, asynchronous delayed invalidation.
//
// Authorization tuple writes are idempotent; therefore a failure
// during the synchronous cache invalidation indicates a dirty cache, but the
// caller can safely retry the operation without causing data corruption.
func (w *TupleWriter) Write(ctx context.Context, tuples []domain.Tuple) error {
	if len(tuples) == 0 {
		return nil
	}

	ctx, span := w.tracer.Start(ctx, "TupleWriter/Write")
	defer span.End()

	span.SetAttributes(
		attribute.Int("authz.tuples_count", len(tuples)),
		attribute.String("authz.first_tuple_user", tuples[0].User),
	)

	keys := make([]string, 0, 2*len(tuples))
	for _, tuple := range tuples {
		keys = append(keys, tuple.User, tuple.Object)
	}

	slices.Sort(keys)
	keys = slices.Compact(keys)

	if err := w.writer.Write(ctx, tuples); err != nil {
		w.logger.ErrorContext(ctx, "failed to land tuples", slog.Any("error", err))
		span.SetStatus(codes.Error, "failed to write tuples")
		span.RecordError(err)

		return fmt.Errorf("write tuples: %w", err)
	}

	// Synchronous invalidation guarantees strong consistency. A failure here
	// means the cache is dirty, triggering an error return so the caller can
	// safely retry the idempotent write operation.
	if err := w.cache.Invalidate(ctx, keys); err != nil {
		w.logger.ErrorContext(ctx, "failed to clean cache (sync)", slog.Any("error", err))
		span.SetStatus(codes.Error, "sync cache invalidation failed")
		span.RecordError(err)

		return fmt.Errorf("invalidate cache: %w", err)
	}

	link := trace.LinkFromContext(ctx)

	// Best-effort enqueue of the async delayed invalidation task.
	// If the queue is full, the task is dropped to prevent blocking the request.
	w.mu.RLock()
	defer w.mu.RUnlock()

	if w.stopped {
		w.logger.WarnContext(ctx, "writer stopped, dropping async invalidation task")
		return nil
	}

	select {
	case w.ch <- &invalidationTask{link: link, keys: keys}:
	default:
		w.logger.WarnContext(ctx, "async invalidation channel full, dropping task")
	}

	w.logger.InfoContext(ctx, "tuples settled", slog.Int("count", len(tuples)))

	return nil
}

// Stop gracefully shuts down the worker pool.
//
// It stops accepting new tasks and immediately drops any pending tasks in the queue to ensure a fast shutdown.
// Because synchronous invalidation already guarantees strong consistency, dropping async tasks is safe.
func (w *TupleWriter) Stop(ctx context.Context) error {
	w.mu.Lock()
	if w.stopped {
		w.mu.Unlock()
		return nil
	}

	w.stopped = true
	close(w.stop)
	w.mu.Unlock()

	done := make(chan struct{})

	go func() {
		w.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return fmt.Errorf("tuple writer graceful shutdown timed out: %w", ctx.Err())
	}
}

func (w *TupleWriter) drainChannel() {
	for {
		select {
		case <-w.ch:
			// discard tasks
		default:
			return
		}
	}
}

// runDelayedInvalidation is the main worker loop.
//
// It listens for incoming tasks, waits for the configured delay, and processes them.
// It responds to the stop signal immediately, even while waiting for a delay.
func (w *TupleWriter) runDelayedInvalidation() {
	defer w.wg.Done()

	timer := time.NewTimer(cacheInvalidationDelay)
	defer timer.Stop()

	for {
		select {
		case task := <-w.ch:
			w.processTask(task, timer)

		case <-w.stop:
			w.drainChannel()
			return
		}
	}
}

func (w *TupleWriter) processTask(task *invalidationTask, timer *time.Timer) {
	if !timer.Stop() {
		<-timer.C
	}

	timer.Reset(cacheInvalidationDelay)

	select {
	case <-timer.C:
		ctx, cancel := context.WithTimeout(context.Background(), cacheInvalidationTimeout)
		defer cancel()

		ctx, span := w.tracer.Start(ctx, "TupleWriter/AsyncCacheInvalidate",
			trace.WithAttributes(attribute.StringSlice("cache.keys", task.keys)),
			trace.WithLinks(task.link),
		)
		defer span.End()

		if err := w.cache.Invalidate(ctx, task.keys); err != nil {
			w.logger.ErrorContext(ctx, "failed to clean cache (async)", slog.Any("error", err))
			span.SetStatus(codes.Error, "async cache invalidation failed")
			span.RecordError(err)
		}

	case <-w.stop:
		w.drainChannel()
		return
	}
}

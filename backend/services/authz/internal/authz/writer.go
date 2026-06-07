package authz

import (
	"context"
	"fmt"
	"log/slog"

	sdk "github.com/openfga/go-sdk/client"
)

// SDKWriter is the subset of the OpenFGA Go SDK client that Writer needs.
type SDKWriter interface {
	Write(ctx context.Context) sdk.SdkClientWriteRequestInterface
}

// Tuple represents an authorization relationship between a subject and an object.
type Tuple struct {
	Subject  string
	Relation string
	Object   string
}

// Writer handles writing authorization tuples to OpenFGA.
type Writer struct {
	client SDKWriter
	logger *slog.Logger
}

// NewWriter creates a Writer backed by the given OpenFGA SDK client.
func NewWriter(client SDKWriter, logger *slog.Logger) *Writer {
	return &Writer{
		client: client,
		logger: logger,
	}
}

// WriteTuples writes multiple authorization tuples to OpenFGA in a single request.
func (w *Writer) WriteTuples(ctx context.Context, tuples []Tuple) error {
	if len(tuples) == 0 {
		return nil
	}

	sdkTuples := make([]sdk.ClientTupleKey, 0, len(tuples))
	for _, t := range tuples {
		sdkTuples = append(sdkTuples, sdk.ClientTupleKey{
			User:     t.Subject,
			Relation: t.Relation,
			Object:   t.Object,
		})
	}

	body := sdk.ClientWriteRequest{
		Writes: sdkTuples,
	}

	_, err := w.client.Write(ctx).Body(body).Execute()
	if err != nil {
		w.logger.ErrorContext(ctx, "failed to write tuples to openfga",
			slog.Int("tuple_count", len(tuples)),
			slog.Any("error", err),
		)

		return fmt.Errorf("openfga write: %w", err)
	}

	w.logger.DebugContext(ctx, "successfully wrote tuples",
		slog.Int("count", len(tuples)),
	)

	return nil
}

// Ensure *OpenFgaClient satisfies goSdkWriter.
var _ SDKWriter = (*sdk.OpenFgaClient)(nil)

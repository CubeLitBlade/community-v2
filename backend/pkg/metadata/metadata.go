// Package metadata provides helpers to extract identity/context values from gRPC metadata.
package metadata

import (
	"context"
	"fmt"
	"strconv"

	"google.golang.org/grpc/metadata"
)

// AccountIDKey is the gRPC metadata key for the authenticated account ID.
const AccountIDKey = "x-account-id"

// AccountIDFromContext extracts and parses the account ID from incoming gRPC metadata.
func AccountIDFromContext(ctx context.Context) (int64, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return 0, fmt.Errorf("missing gRPC metadata")
	}
	vals := md.Get(AccountIDKey)
	if len(vals) == 0 {
		return 0, fmt.Errorf("missing metadata key %q", AccountIDKey)
	}
	return strconv.ParseInt(vals[0], 10, 64)
}
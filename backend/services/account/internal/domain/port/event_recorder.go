package port

import (
	"context"

	"github.com/cubelitblade/community-v2/backend/pkg/events/outbox"
)

type EventRecorder interface {
	Record(ctx context.Context, entry *outbox.Entry) error
}

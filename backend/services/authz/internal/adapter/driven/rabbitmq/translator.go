package rabbitmq

import (
	"errors"
	"fmt"

	"github.com/cloudevents/sdk-go/v2/event"

	sharedevents "github.com/cubelitblade/community-v2/backend/pkg/events/account/v1"
	postevents "github.com/cubelitblade/community-v2/backend/pkg/events/post/v1"
	"github.com/cubelitblade/community-v2/backend/services/authz/internal/domain"
	"github.com/cubelitblade/community-v2/backend/services/authz/permission"
)

var ErrUnknownEvent = errors.New("unknown event")

func translate(evt event.Event) (domain.Event, error) {
	switch evt.Type() {
	case sharedevents.EventTypeAccountCreated:
		return translateAccountCreated(evt)
	case postevents.EventTypePostPublished:
		return translatePostPublished(evt)
	default:
		return nil, fmt.Errorf("%w: %s", ErrUnknownEvent, evt.Type())
	}
}

func translateAccountCreated(evt event.Event) (*domain.AccountCreated, error) {
	var payload sharedevents.AccountCreated
	if err := evt.DataAs(&payload); err != nil {
		return nil, fmt.Errorf("unmarshal account created data: %w", err)
	}

	return &domain.AccountCreated{
		Tuples: []domain.Tuple{
			{
				User:     "user:" + payload.AccountId,
				Relation: payload.Role,
				Object:   permission.ObjectSystemCommunity,
			},
		},
	}, nil
}

func translatePostPublished(evt event.Event) (*domain.PostPublished, error) {
	var payload postevents.PostPublished
	if err := evt.DataAs(&payload); err != nil {
		return nil, fmt.Errorf("unmarshal post published data: %w", err)
	}

	return &domain.PostPublished{
		Tuples: []domain.Tuple{
			{
				User:     "user:" + payload.AuthorID,
				Relation: "author",
				Object:   "post:" + payload.PostID,
			},
			{
				User:     permission.ObjectSystemCommunity,
				Relation: "global",
				Object:   "post:" + payload.PostID,
			},
		},
	}, nil
}

package rabbitmq

import (
	"errors"
	"fmt"

	"github.com/cloudevents/sdk-go/v2/event"

	sharedevents "github.com/cubelitblade/community-v2/backend/pkg/events/account/v1"
	"github.com/cubelitblade/community-v2/backend/services/authz/internal/domain"
	"github.com/cubelitblade/community-v2/backend/services/authz/permission"
)

var ErrUnknownEvent = errors.New("unknown event")

func translate(evt event.Event) (domain.Event, error) {
	switch evt.Type() {
	case sharedevents.EventTypeAccountCreated:
		return translateAccountCreated(evt)
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

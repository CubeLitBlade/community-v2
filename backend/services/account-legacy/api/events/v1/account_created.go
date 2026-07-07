package v1

import sharedevents "github.com/cubelitblade/community-v2/backend/pkg/events/account/v1"

const (
	TopicAccountCreated     = sharedevents.TopicAccountCreated
	EventTypeAccountCreated = sharedevents.EventTypeAccountCreated

	AggregateTypeAccountService = sharedevents.AggregateType
)

type AccountCreatedEventPayload = sharedevents.AccountCreated

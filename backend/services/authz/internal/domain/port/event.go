package port

import (
	"context"
	"errors"

	"github.com/cloudevents/sdk-go/v2/event"
)

type EventHandler interface {
	Handle(ctx context.Context, evt event.Event) error
}

var ErrPoisonMessage = errors.New("poison message")

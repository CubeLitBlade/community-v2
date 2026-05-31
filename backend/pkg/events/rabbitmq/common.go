package rabbitmq

import (
	"errors"
	"time"

	rmq "github.com/rabbitmq/rabbitmq-amqp-go-client/pkg/rabbitmqamqp"
)

// Default configuration values for publishers and subscribers.
const (
	DefaultCloseTimeout   = 30 * time.Second
	DefaultInitialCredits = 10
)

var (
	// ErrMessageStateRejected indicates that the published message was rejected by the broker.
	ErrMessageStateRejected = errors.New("message was rejected")

	// ErrMessageStateReleased indicates that the published message was released by the broker.
	ErrMessageStateReleased = errors.New("message was released")

	// ErrMessageStateModified indicates that the published message was modified by the broker.
	ErrMessageStateModified = errors.New("message was modified")

	// ErrMessageStateUnknown indicates that the broker returned an unknown message outcome state.
	ErrMessageStateUnknown = errors.New("message state unknown")

	// ErrPublisherClosed indicates that a publish operation was attempted while the publisher is closed.
	ErrPublisherClosed = errors.New("publisher is closed")
)

// CheckOutcome evaluates an AMQP delivery state and returns an error if the outcome is not Accepted.
func CheckOutcome(outcome rmq.DeliveryState) error {
	switch outcome.(type) {
	case *rmq.StateAccepted:
		return nil
	case *rmq.StateRejected:
		return ErrMessageStateRejected
	case *rmq.StateReleased:
		return ErrMessageStateReleased
	case *rmq.StateModified:
		return ErrMessageStateModified
	default:
		return ErrMessageStateUnknown
	}
}

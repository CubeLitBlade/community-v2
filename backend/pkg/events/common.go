package events

import (
	"errors"
	"time"
)

const (
	defaultCloseTimeout   = 30 * time.Second
	defaultInitialCredits = 10
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

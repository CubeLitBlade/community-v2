// Package events provides abstractions for publishing and consuming messages
// via RabbitMQ using the AMQP protocol.
package events

import "github.com/Azure/go-amqp"

// ReceivedMessage represents a message consumed from a RabbitMQ queue,
// providing methods to acknowledge, requeue, or discard it.
type ReceivedMessage struct {
	// Data holds the byte payload of the received message.
	Data []byte

	acceptFunc  func() error
	requeueFunc func() error
	discardFunc func(e *amqp.Error) error
}

// Accept indicates that the message was successfully processed.
// RabbitMQ will delete the message.
func (m *ReceivedMessage) Accept() error {
	if m.acceptFunc != nil {
		return m.acceptFunc()
	}
	return nil
}

// Requeue indicates that the message was not processed.
// RabbitMQ will requeue the message for redelivery to the same or a different consumer.
func (m *ReceivedMessage) Requeue() error {
	if m.requeueFunc != nil {
		return m.requeueFunc()
	}
	return nil
}

// Discard indicates that the message is invalid and unprocessable.
// RabbitMQ will dead-letter or drop the message without re-queuing,
// optionally using the provided AMQP error.
func (m *ReceivedMessage) Discard(e *amqp.Error) error {
	if m.discardFunc != nil {
		return m.discardFunc(e)
	}
	return nil
}

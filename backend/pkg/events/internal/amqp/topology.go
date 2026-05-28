package amqp

import (
	"context"
	"errors"
	"fmt"

	rmq "github.com/rabbitmq/rabbitmq-amqp-go-client/pkg/rabbitmqamqp"
)

func ConnectWithTopicExchange(
	ctx context.Context, env *rmq.Environment, exchangeName string,
) (*rmq.AmqpConnection, error) {
	conn, err := Connect(ctx, env)
	if err != nil {
		return nil, fmt.Errorf("connect AMQP connection: %w", err)
	}

	if err := DeclareTopicExchange(ctx, conn, exchangeName); err != nil {
		err = fmt.Errorf("declare topic exchange: %w", err)

		if closeErr := conn.Close(ctx); closeErr != nil {
			return nil, errors.Join(err, fmt.Errorf("close AMQP connection: %w", closeErr))
		}

		return nil, err
	}

	return conn, nil
}

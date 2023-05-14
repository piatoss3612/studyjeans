package tools

import (
	"context"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

func DialRabbitMQ(addr string) (*amqp.Connection, error) {
	conn, err := amqp.Dial(addr)
	if err != nil {
		return nil, err
	}

	return conn, nil
}

func RedialRabbitMQ(ctx context.Context, addr string) <-chan *amqp.Connection {
	ch := make(chan *amqp.Connection)

	go func() {
		defer close(ch)

		for {
			select {
			case <-ctx.Done():
				return
			default:
				conn, err := DialRabbitMQ(addr)
				if err != nil {
					time.Sleep(500 * time.Millisecond)
					continue
				}

				ch <- conn
				return
			}
		}
	}()

	return ch
}

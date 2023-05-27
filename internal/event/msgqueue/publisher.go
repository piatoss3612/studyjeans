package msgqueue

import (
	"context"
	"encoding/json"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Publisher interface {
	Publish(ctx context.Context, k string, v any) error
}

type publisherImpl struct {
	conn     *amqp.Connection
	exchange string
}

func NewPublisher(conn *amqp.Connection, exchange, kind string) (Publisher, error) {
	pub := &publisherImpl{
		conn:     conn,
		exchange: exchange,
	}

	return pub.setup(exchange, kind)
}

func (p *publisherImpl) setup(exchange, kind string) (Publisher, error) {
	ch, err := p.conn.Channel()
	if err != nil {
		return nil, err
	}
	defer func() { _ = ch.Close() }()

	err = ch.ExchangeDeclare(exchange, kind, true, false, false, false, nil)
	if err != nil {
		return nil, err
	}

	return p, nil
}

func (p *publisherImpl) Publish(ctx context.Context, k string, v any) error {
	ch, err := p.conn.Channel()
	if err != nil {
		return err
	}
	defer func() { _ = ch.Close() }()

	body, err := json.Marshal(v)
	if err != nil {
		return err
	}

	msg := amqp.Publishing{
		Headers: amqp.Table{
			"x-event-name": k,
		},
		ContentType: "application/json",
		Body:        body,
	}

	return ch.PublishWithContext(ctx, p.exchange, k, false, false, msg)
}

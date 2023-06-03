package rabbitmq

import (
	"context"
	"encoding/json"

	"github.com/piatoss3612/my-study-bot/internal/pubsub"
	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	XEventTopicHeader = "x-event-topic"
	ContentTypeJson   = "application/json"
)

type publisher struct {
	conn     *amqp.Connection
	exchange string
}

func NewPublisher(conn *amqp.Connection, exchange, kind string) (pubsub.Publisher, error) {
	pub := &publisher{
		conn:     conn,
		exchange: exchange,
	}

	return pub.setup(exchange, kind)
}

func (p *publisher) setup(exchange, kind string) (pubsub.Publisher, error) {
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

func (p *publisher) Publish(ctx context.Context, k string, v any) error {
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
			XEventTopicHeader: k,
		},
		ContentType: ContentTypeJson,
		Body:        body,
	}

	return ch.PublishWithContext(ctx, p.exchange, k, false, false, msg)
}

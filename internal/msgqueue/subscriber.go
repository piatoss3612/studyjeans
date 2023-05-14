package msgqueue

import (
	"context"
	"errors"

	amqp "github.com/rabbitmq/amqp091-go"
)

var ErrMissingEventNameHeader = errors.New("missing x-event-name header")

type Message struct {
	EventName string
	Body      []byte
}

type Subscriber interface {
	Subscribe(ctx context.Context, topics ...string) (<-chan Message, <-chan error, error)
}

type subscriberImpl struct {
	conn     *amqp.Connection
	exchange string
	queue    string
}

func NewSubscriber(conn *amqp.Connection, exchange, kind, queue string) (Subscriber, error) {
	sub := &subscriberImpl{
		conn:     conn,
		exchange: exchange,
		queue:    queue,
	}

	return sub.setup(exchange, kind, queue)
}

func (s *subscriberImpl) setup(exchange, kind, queue string) (Subscriber, error) {
	ch, err := s.conn.Channel()
	if err != nil {
		return nil, err
	}
	defer func() { _ = ch.Close() }()

	err = ch.ExchangeDeclare(exchange, kind, true, false, false, false, nil)
	if err != nil {
		return nil, err
	}

	_, err = ch.QueueDeclare(queue, true, false, false, false, nil)
	if err != nil {
		return nil, err
	}

	return s, nil
}

func (s *subscriberImpl) Subscribe(ctx context.Context, topics ...string) (<-chan Message, <-chan error, error) {
	ch, err := s.conn.Channel()
	if err != nil {
		return nil, nil, err
	}

	for _, topic := range topics {
		if err := ch.QueueBind(s.queue, topic, s.exchange, false, nil); err != nil {
			return nil, nil, err
		}
	}

	delivery, err := ch.Consume(s.queue, "", false, false, false, false, nil)
	if err != nil {
		return nil, nil, err
	}

	messages := make(chan Message)
	errors := make(chan error)

	go s.handleMessage(ctx, delivery, messages, errors)

	return messages, errors, nil
}

func (s *subscriberImpl) handleMessage(ctx context.Context, delivery <-chan amqp.Delivery, msgs chan Message, errs chan error) {
	defer func() {
		close(msgs)
		close(errs)
	}()

	select {
	case <-ctx.Done():
		return
	case d := <-delivery:
		eventName, ok := d.Headers["x-event-name"]
		if !ok {
			errs <- ErrMissingEventNameHeader
			_ = d.Ack(false)
			break
		}

		msg := Message{
			EventName: eventName.(string),
			Body:      d.Body,
		}

		msgs <- msg
		_ = d.Ack(false)
	}
}

package pubsub

import (
	"context"
)

type Publisher interface {
	Publish(ctx context.Context, k string, v any) error
}

type Subscriber interface {
	Subscribe(topics ...string) (<-chan Message, <-chan error, func(), error)
}

type Mapper interface {
	RegisterHandler(topic string, handler Handler)
	Map(topic string) (Handler, bool)
}

type Handler interface {
	Handle(body []byte) error
}

type Message struct {
	Topic string
	Body  []byte
}

type mapper struct {
	handlers map[string]Handler
}

func NewMapper() Mapper {
	return &mapper{
		handlers: make(map[string]Handler),
	}
}

func (m *mapper) RegisterHandler(topic string, handler Handler) {
	m.handlers[topic] = handler
}

func (m *mapper) Map(topic string) (Handler, bool) {
	handler, ok := m.handlers[topic]
	return handler, ok
}

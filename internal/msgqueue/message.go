package msgqueue

type Message struct {
	EventName string
	Body      []byte
}

type Mapper interface {
	RegisterHandler(topic string, handler Handler)
	Map(topic string) (Handler, bool)
}

type Handler interface {
	Handle(msg Message) error
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

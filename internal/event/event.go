package event

type Event interface {
	Topic() string
	Description() string
	Timestamp() int64
}

package event

type Event interface {
	Name() string
	Description() string
	Timestamp() int64
}

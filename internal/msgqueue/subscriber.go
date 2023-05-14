package msgqueue

type Subscriber interface {
	Subscribe(topics ...string) (<-chan []byte, <-chan error, error)
}

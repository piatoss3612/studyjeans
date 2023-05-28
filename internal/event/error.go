package event

import "time"

type ErrorEvent struct {
	T string    `json:"topic"`
	D string    `json:"description"`
	C time.Time `json:"created_at"`
}

func (e *ErrorEvent) Topic() string {
	return e.T
}

func (e *ErrorEvent) Description() string {
	return e.D
}

func (e *ErrorEvent) Timestamp() int64 {
	return e.C.Unix()
}

var _ = Event(&ErrorEvent{})

package event

import (
	"time"
)

type StudyEvent struct {
	T string    `json:"topic"`
	D string    `json:"description"`
	C time.Time `json:"created_at"`
}

func (s StudyEvent) Topic() string {
	return s.T
}

func (s StudyEvent) Description() string {
	return s.D
}

func (s StudyEvent) Timestamp() int64 {
	return s.C.Unix()
}

var _ = Event(&StudyEvent{})

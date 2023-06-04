package errors

import (
	"errors"
	"time"
)

type EventTopic string

var EventTopicError EventTopic = "error"

var ErrUnknownEventTopic = errors.New("unknown event topic")

func (t EventTopic) Validate() error {
	switch t {
	case EventTopicError:
	default:
		return ErrUnknownEventTopic
	}

	return nil
}

func (t EventTopic) String() string {
	return string(t)
}

type Event struct {
	Topic       EventTopic `json:"topic"`
	Description string     `json:"description"`
	Timestamp   int64      `json:"timestamp"`
}

func NewEvent(topic EventTopic, description string) (Event, error) {
	if err := topic.Validate(); err != nil {
		return Event{}, err
	}

	evt := Event{
		Topic:       topic,
		Description: description,
		Timestamp:   time.Now().Unix(),
	}

	return evt, nil
}

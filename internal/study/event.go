package study

import "time"

type EventTopic string

var (
	EventTopicStudyRoundCreated  EventTopic = "study.round.created"
	EventTopicStudyRoundMoved    EventTopic = "study.round.moved"
	EventTopicStudyRoundFinished EventTopic = "study.round.finished"
)

func (t EventTopic) Validate() error {
	switch t {
	case EventTopicStudyRoundCreated, EventTopicStudyRoundMoved, EventTopicStudyRoundFinished:
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
	Data        any        `json:"data"`
}

func NewEvent(topic EventTopic, description string, data ...any) (Event, error) {
	if err := topic.Validate(); err != nil {
		return Event{}, err
	}

	evt := Event{
		Topic:       topic,
		Description: description,
		Timestamp:   time.Now().Unix(),
	}

	if len(data) > 0 {
		evt.Data = data[0]
	}

	return evt, nil
}

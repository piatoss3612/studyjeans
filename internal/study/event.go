package study

import "time"

type EventTopic string

var (
	EventTopicStudyRoundCreated  EventTopic = "study.round.created"
	EventTopicStudyRoundProgress EventTopic = "study.round.progress"
	EventTopicStudyRoundFinished EventTopic = "study.round.finished"
)

func (t EventTopic) Validate() error {
	switch t {
	case EventTopicStudyRoundCreated, EventTopicStudyRoundProgress, EventTopicStudyRoundFinished:
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
	Data        []byte     `json:"data"`
}

func NewEvent(topic EventTopic, description string, data ...[]byte) (Event, error) {
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

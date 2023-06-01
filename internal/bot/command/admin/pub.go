package admin

import (
	"context"
	"time"

	"github.com/piatoss3612/presentation-helper-bot/internal/event"
	"github.com/piatoss3612/presentation-helper-bot/internal/study"
)

// publish round to topic
func (ac *adminCommand) publishRound(topic string, round *study.Round) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	for i := 0; i < 10; i++ {
		err := ac.pub.Publish(ctx, topic, round)
		if err != nil {
			ac.sugar.Errorw(err.Error(), "event", "publish round", "topic", topic, "try", i+1)
			continue
		}
		return
	}
}

// publish event to event topic
func (ac *adminCommand) publishEvent(evt event.Event) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	for i := 0; i < 10; i++ {
		err := ac.pub.Publish(ctx, evt.Topic(), evt)
		if err != nil {
			ac.sugar.Errorw(err.Error(), "event", "publish event", "topic", evt.Topic(), "try", i+1)
			continue
		}
		return
	}
}

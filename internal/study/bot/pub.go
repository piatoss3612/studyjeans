package bot

import (
	"context"
	"time"

	"github.com/piatoss3612/presentation-helper-bot/internal/event"
	"github.com/piatoss3612/presentation-helper-bot/internal/study"
)

func (b *StudyBot) publishRound(topic string, round *study.Round) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	for i := 0; i < 10; i++ {
		err := b.pub.Publish(ctx, topic, round)
		if err != nil {
			b.sugar.Errorw(err.Error(), "event", "publish round", "topic", topic, "try", i+1)
			continue
		}
		return
	}
}

func (b *StudyBot) publishEvent(evt event.Event) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	for i := 0; i < 10; i++ {
		err := b.pub.Publish(ctx, evt.Topic(), evt)
		if err != nil {
			b.sugar.Errorw(err.Error(), "event", "publish event", "topic", evt.Topic(), "try", i+1)
			continue
		}
		return
	}
}

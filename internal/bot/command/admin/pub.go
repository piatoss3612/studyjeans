package admin

import (
	"context"
	"time"

	"github.com/piatoss3612/my-study-bot/internal/event"
	"github.com/piatoss3612/my-study-bot/internal/study"
)

// publish round info
func (ac *adminCommand) publishRoundOnRoundClosed(round *study.Round) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cnt := 0
	topic := "study.round-closed"

	for {
		select {
		case <-ctx.Done():
			ac.sugar.Errorw("failed to publish round info", "error", ctx.Err().Error(), "topic", topic, "retry", cnt)
			return
		default:
			err := ac.pub.Publish(ctx, topic, round)
			if err != nil {
				ac.sugar.Errorw("failed to publish round", "error", err.Error(), "topic", topic, "retry", cnt)
				time.Sleep(500 * time.Millisecond)
				cnt++
				continue
			}
			ac.sugar.Infow("round published", "topic", topic, "retry", cnt)
			return
		}
	}
}

// publish event to event topic
func (ac *adminCommand) publishRoundProgress(evt event.Event) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cnt := 0

	for {
		select {
		case <-ctx.Done():
			ac.sugar.Errorw("failed to publish event", "error", ctx.Err().Error(), "topic", evt.Topic(), "description", evt.Description(), "retry", cnt)
			return
		default:
			err := ac.pub.Publish(ctx, evt.Topic(), evt)
			if err != nil {
				ac.sugar.Errorw("failed to publish event", "error", err.Error(), "topic", evt.Topic(), "description", evt.Description(), "retry", cnt)
				time.Sleep(500 * time.Millisecond)
				cnt++
				continue
			}
			ac.sugar.Infow("event published", "topic", evt.Topic(), "retry", cnt)
			return
		}
	}
}

package service

import (
	"context"
	"time"
)

func (l *LoggerService) Listen(stop <-chan bool, topics []string) {
	msgs, errs, close, err := l.sub.Subscribe(topics...)
	if err != nil {
		l.sugar.Fatal(err)
	}
	defer close()

	for {
		select {
		case msg := <-msgs:
			h, ok := l.mapper.Map(msg.Topic)
			if !ok {
				l.sugar.Errorw("Unknown event name", "event", msg.Topic)
				continue
			}

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			if err := h.Handle(ctx, msg.Body); err != nil {
				l.sugar.Errorw("Failed to handle event", "event", msg.Topic, "error", err)
				// TODO: retry?
				continue
			}
			l.sugar.Infow("Successfully handled event", "event", msg.Topic)
		case err := <-errs:
			if err == nil {
				continue
			}
			l.sugar.Errorw("Received error from subscriber", "error", err)
		case <-stop:
			l.sugar.Info("Stop listening to events")
			return
		default:
			time.Sleep(500 * time.Millisecond)
		}
	}
}

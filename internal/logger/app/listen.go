package app

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/piatoss3612/presentation-helper-bot/internal/logger"
)

func (l *LoggerApp) Listen(stop <-chan bool, topics []string) {
	msgs, errs, close, err := l.sub.Subscribe(topics...)
	if err != nil {
		l.sugar.Fatal(err)
	}
	defer close()

	for {
		select {
		case msg := <-msgs:
			fields := strings.Split(msg.EventName, ".")
			if len(fields) != 2 {
				l.sugar.Errorf("Invalid event name", "event", msg.EventName)
				continue
			}

			switch fields[0] {
			case "study":
				switch fields[1] {
				case "closed":
					round := logger.NewRound()
					if err := json.Unmarshal(msg.Body, &round); err != nil {
						l.sugar.Errorf("Failed to unmarshal message body", "error", err)
						continue
					}

					ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
					defer cancel()

					err := l.svc.RecordRound(ctx, round)
					if err != nil {
						l.sugar.Errorf("Failed to record round", "error", err)
					}
				default:
					// TODO
				}
			default:
				l.sugar.Errorf("Unknown event name", "event", msg.EventName)
			}
		case err := <-errs:
			if err == nil {
				continue
			}
		case <-stop:
			return
		default:
			time.Sleep(100 * time.Millisecond)
		}
	}
}

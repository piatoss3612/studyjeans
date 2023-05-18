package recorder

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	models "github.com/piatoss3612/presentation-helper-bot/internal/models/recorder"
)

func (r *Recorder) Listen(stop <-chan bool, topics []string) {
	msgs, errs, close, err := r.sub.Subscribe(topics...)
	if err != nil {
		r.sugar.Fatal(err)
	}
	defer close()

	for {
		select {
		case msg := <-msgs:
			fields := strings.Split(msg.EventName, ".")
			if len(fields) != 2 {
				r.sugar.Errorf("Invalid event name", "event", msg.EventName)
				continue
			}

			switch fields[0] {
			case "study":
				switch fields[1] {
				case "closed":
					round := models.NewRound()
					if err := json.Unmarshal(msg.Body, &round); err != nil {
						r.sugar.Errorf("Failed to unmarshal message body", "error", err)
						continue
					}

					ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
					defer cancel()

					err := r.svc.RecordRound(ctx, round)
					if err != nil {
						r.sugar.Errorf("Failed to record round", "error", err)
					}
				default:
					// TODO
				}
			default:
				r.sugar.Errorf("Unknown event name", "event", msg.EventName)
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

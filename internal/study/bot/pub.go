package bot

import (
	"context"
	"time"

	"github.com/piatoss3612/presentation-helper-bot/internal/study"
)

func (b *StudyBot) publishRoundInfo(round *study.Round) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	for i := 0; i < 10; i++ {
		err := b.pub.Publish(ctx, "study.closed", round)
		if err != nil {
			b.sugar.Errorw(err.Error(), "topic", "study.finished")
			continue
		}
		return
	}
}

package study

import (
	"context"
	"time"

	"github.com/piatoss3612/presentation-helper-bot/internal/models/study"
)

func (b *StudyBot) setRound(ctx context.Context, s *study.Round) error {
	return b.cache.Set(ctx, s.GuildID, s, 3*time.Minute)
}

func (b *StudyBot) setRoundRetry(r *study.Round, timeout time.Duration) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	for {
		err := b.setRound(ctx, r)
		if err != nil {
			b.sugar.Errorw(err.Error(), "event", "study-round-info")

			if err == context.DeadlineExceeded {
				return
			}

			time.Sleep(500 * time.Millisecond)
			continue
		}
		break
	}
}

func (b *StudyBot) getRound(ctx context.Context, guildID string) (*study.Round, error) {
	var round study.Round

	if err := b.cache.Get(ctx, guildID, &round); err != nil {
		return nil, err
	}

	return &round, nil
}

func (b *StudyBot) roundExists(ctx context.Context, guildID string) bool {
	return b.cache.Exists(ctx, guildID)
}

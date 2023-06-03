package info

import (
	"context"
	"time"

	"github.com/piatoss3612/my-study-bot/internal/study"
)

func (ic *infoCommand) setRound(ctx context.Context, s *study.Round) error {
	return ic.cache.Set(ctx, s.GuildID, s, 3*time.Minute)
}

func (ic *infoCommand) getRound(ctx context.Context, guildID string) (*study.Round, error) {
	var round study.Round

	if err := ic.cache.Get(ctx, guildID, &round); err != nil {
		return nil, err
	}

	return &round, nil
}

func (ic *infoCommand) roundExists(ctx context.Context, guildID string) bool {
	return ic.cache.Exists(ctx, guildID)
}

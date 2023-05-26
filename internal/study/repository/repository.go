package repository

import (
	"context"

	"github.com/piatoss3612/presentation-helper-bot/internal/study"
)

type Query interface {
	FindStudy(ctx context.Context, guildID string) (*study.Study, error)
	FindRound(ctx context.Context, roundID string) (*study.Round, error)
	FindRounds(ctx context.Context, guildID string) ([]*study.Round, error)
}

type Store interface {
	CreateStudy(ctx context.Context, s study.Study) (*study.Study, error)
	UpdateStudy(ctx context.Context, s study.Study) (*study.Study, error)
	CreateRound(ctx context.Context, r study.Round) (*study.Round, error)
	UpdateRound(ctx context.Context, r study.Round) (*study.Round, error)
}

type Tx interface {
	Store
	Query
	ExecTx(ctx context.Context, fn func(ctx context.Context) (interface{}, error)) (interface{}, error)
}

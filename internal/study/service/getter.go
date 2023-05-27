package study

import (
	"context"

	"github.com/piatoss3612/presentation-helper-bot/internal/study"
)

// get study of guild
func (svc *serviceImpl) GetStudy(ctx context.Context, guildID string) (*study.Study, error) {
	defer svc.mtx.Unlock()
	svc.mtx.Lock()

	s, err := svc.tx.FindStudy(ctx, guildID)
	if err != nil {
		return nil, err
	}

	if s == nil {
		return nil, study.ErrStudyNotFound
	}

	return s, nil
}

// get ongoing round of study of guild
func (svc *serviceImpl) GetOngoingRound(ctx context.Context, guildID string) (*study.Round, error) {
	defer svc.mtx.Unlock()
	svc.mtx.Lock()

	txFn := func(sc context.Context) (interface{}, error) {
		// find study
		s, err := svc.tx.FindStudy(sc, guildID)
		if err != nil {
			return nil, err
		}

		// if there is no study, return error
		if s == nil {
			return nil, study.ErrStudyNotFound
		}

		// if there is no ongoing round, return error
		if s.OngoingRoundID == "" {
			return nil, study.ErrRoundNotFound
		}

		// find ongoing round
		r, err := svc.tx.FindRound(sc, s.OngoingRoundID)
		if err != nil {
			return nil, err
		}

		// if there is no ongoing round, return error
		if r == nil {
			return nil, study.ErrStudyNotFound
		}

		return r, nil
	}

	// execute transaction
	res, err := svc.tx.ExecTx(ctx, txFn)
	if err != nil {
		return nil, err
	}

	return res.(*study.Round), nil
}

// get all rounds of study of guild
func (svc *serviceImpl) GetRounds(ctx context.Context, guildID string) ([]*study.Round, error) {
	defer svc.mtx.Unlock()
	svc.mtx.Lock()

	r, err := svc.tx.FindRounds(ctx, guildID)
	if err != nil {
		return nil, err
	}

	if r == nil {
		return nil, study.ErrRoundNotFound
	}

	return r, nil
}

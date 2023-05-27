package service

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/piatoss3612/presentation-helper-bot/internal/study"
	"github.com/piatoss3612/presentation-helper-bot/internal/study/repository"
)

type Service interface {
	NewStudy(ctx context.Context, guildID, managerID string) (*study.Study, error)
	NewRound(ctx context.Context, guildID, title string, memberIDs []string) (*study.Study, error)
	Update(ctx context.Context, params *UpdateParams, updates ...UpdateFunc) (*study.Study, error)
	GetStudy(ctx context.Context, guildID string) (*study.Study, error)
	GetOngoingRound(ctx context.Context, guildID string) (*study.Round, error)
	GetRounds(ctx context.Context, guildID string) ([]*study.Round, error)
}

type UpdateParams struct {
	GuildID    string
	ManagerID  string
	ChannelID  string
	MemberID   string
	MemberName string
	Subject    string
	ContentURL string
	ReviewerID string
	RevieweeID string
}

type UpdateFunc func(*study.Study, *study.Round, *UpdateParams) error

type studyService struct {
	tx repository.Tx

	mtx *sync.Mutex
}

// create new service
func New(ctx context.Context, tx repository.Tx, guildID, managerID, noticeChID, reflectionChID string) Service {
	svc := &studyService{
		tx:  tx,
		mtx: &sync.Mutex{},
	}
	return svc
}

// create new study
func (svc *studyService) NewStudy(ctx context.Context, guildID, managerID string) (*study.Study, error) {
	defer svc.mtx.Unlock()
	svc.mtx.Lock()

	txFn := func(sc context.Context) (interface{}, error) {
		// find study of guild
		s, err := svc.tx.FindStudy(sc, guildID)
		if err != nil {
			return nil, err
		}

		// if study exists, return error
		if s != nil {
			return nil, study.ErrStudyExists
		}

		// create new study
		ns := study.New()

		ns.SetGuildID(guildID)
		ns.SetManagerID(managerID)

		// store new study
		return svc.tx.CreateStudy(sc, ns)
	}

	// execute transaction
	s, err := svc.tx.ExecTx(ctx, txFn)
	if err != nil {
		return nil, err
	}

	// return created study
	return s.(*study.Study), nil
}

// initialize new study round
func (svc *studyService) NewRound(ctx context.Context, guildID, title string, memberIDs []string) (*study.Study, error) {
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

		// check if there is no ongoing round
		if !(s.CurrentStage.IsNone() || s.CurrentStage.IsWait()) {
			return nil, study.ErrStudyExists
		}

		// increment total round and set current stage
		s.IncrementTotalRound()
		s.SetCurrentStage(study.StageRegistrationOpened)

		// create new round
		r := study.NewRound()
		r.SetGuildID(s.GuildID)
		r.SetNumber(s.TotalRound)
		r.SetTitle(title)
		r.SetStage(study.StageRegistrationOpened)

		// set initial members
		for _, id := range memberIDs {
			member := study.NewMember()
			r.SetMember(id, member)
		}

		// store new round
		ur, err := svc.tx.CreateRound(sc, r)
		if err != nil {
			return nil, err
		}

		s.SetOngoingRoundID(ur.ID)

		// update study
		return svc.tx.UpdateStudy(sc, *s)
	}

	// execute transaction
	s, err := svc.tx.ExecTx(ctx, txFn)
	if err != nil {
		return nil, err
	}

	// return updated study
	return s.(*study.Study), nil
}

// update study or round
func (svc *studyService) Update(ctx context.Context, params *UpdateParams, updates ...UpdateFunc) (*study.Study, error) {
	defer svc.mtx.Unlock()
	svc.mtx.Lock()

	if params == nil {
		return nil, study.ErrNilUpdateParams
	}

	txFn := func(sc context.Context) (interface{}, error) {
		s, err := svc.tx.FindStudy(sc, params.GuildID)
		if err != nil {
			return nil, err
		}

		if s == nil {
			return nil, study.ErrStudyNotFound
		}

		r, err := svc.tx.FindRound(sc, s.OngoingRoundID)
		if err != nil {
			return nil, err
		}

		if r == nil {
			return nil, study.ErrRoundNotFound
		}

		// validate update params
		for _, v := range updates {
			if err := v(s, r, params); err != nil {
				return nil, err
			}
		}

		// update study
		s, err = svc.tx.UpdateStudy(sc, *s)
		if err != nil {
			return nil, err
		}

		// update round
		r, err = svc.tx.UpdateRound(sc, *r)
		if err != nil {
			return nil, err
		}

		return s, nil
	}

	// execute transaction
	s, err := svc.tx.ExecTx(ctx, txFn)
	if err != nil {
		return nil, err
	}

	// return updated study
	return s.(*study.Study), nil
}

// close study round
func (svc *studyService) CloseRound(ctx context.Context, guildID string) (*study.Study, error) {
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

		// check if review is finished
		if s.CurrentStage != study.StageReviewClosed {
			return nil, errors.Join(study.ErrInvalidStage, fmt.Errorf("스터디 라운드 %s 종료가 불가능합니다", s.CurrentStage.String()))
		}

		// update management
		s.SetOngoingRoundID("")
		s.SetCurrentStage(study.StageWait)

		// update study
		return svc.tx.UpdateStudy(sc, *s)
	}

	// execute transaction
	s, err := svc.tx.ExecTx(ctx, txFn)
	if err != nil {
		return nil, err
	}
	return s.(*study.Study), nil
}

// get study of guild
func (svc *studyService) GetStudy(ctx context.Context, guildID string) (*study.Study, error) {
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
func (svc *studyService) GetOngoingRound(ctx context.Context, guildID string) (*study.Round, error) {
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
func (svc *studyService) GetRounds(ctx context.Context, guildID string) ([]*study.Round, error) {
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

package service

import (
	"context"
	"sync"

	"github.com/piatoss3612/presentation-helper-bot/internal/study"
	"github.com/piatoss3612/presentation-helper-bot/internal/study/repository"
)

type Service interface {
	GetRound(ctx context.Context, roundID string) (*study.Round, error)
	GetRounds(ctx context.Context, guildID string) ([]*study.Round, error)
	GetStudy(ctx context.Context, guildID string) (*study.Study, error)
	NewRound(ctx context.Context, params *NewRoundParams) (*study.Study, error)
	NewStudy(ctx context.Context, params *NewStudyParams) (*study.Study, error)
	UpdateRound(ctx context.Context, params *UpdateParams, update UpdateFunc, validators ...UpdateValidator) (*study.Study, error)
	UpdateStudy(ctx context.Context, params *UpdateParams, update UpdateFunc, validators ...UpdateValidator) (*study.Study, error)
}

type studyService struct {
	tx repository.Tx

	mtx *sync.Mutex
}

// create new service
func New(tx repository.Tx) Service {
	svc := &studyService{
		tx:  tx,
		mtx: &sync.Mutex{},
	}
	return svc
}

type NewRoundParams struct {
	GuildID   string
	ManagerID string
	Title     string
	MemberIDs []string
}

type NewStudyParams struct {
	GuildID             string
	ManagerID           string
	NoticeChannelID     string
	ReflectionChannelID string
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

type UpdateFunc func(*study.Study, *study.Round, *UpdateParams)
type UpdateValidator func(*study.Study, *study.Round, *UpdateParams) error

// get round by id
func (svc *studyService) GetRound(ctx context.Context, roundID string) (*study.Round, error) {
	defer svc.mtx.Unlock()
	svc.mtx.Lock()

	txFn := func(sc context.Context) (interface{}, error) {
		// find ongoing round
		r, err := svc.tx.FindRound(sc, roundID)
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

// get all rounds of study by guild id
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

// get study by guild id
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

// initialize new study round
func (svc *studyService) NewRound(ctx context.Context, params *NewRoundParams) (*study.Study, error) {
	defer svc.mtx.Unlock()
	svc.mtx.Lock()

	if params == nil {
		return nil, study.ErrNilParams
	}

	txFn := func(sc context.Context) (interface{}, error) {
		// find study
		s, err := svc.tx.FindStudy(sc, params.GuildID)
		if err != nil {
			return nil, err
		}

		// if there is no study, return error
		if s == nil {
			return nil, study.ErrStudyNotFound
		}

		// check if manager is the one who requested
		if !s.IsManager(params.ManagerID) {
			return nil, study.ErrInvalidManager
		}

		// check if there is any ongoing round
		if !(s.CurrentStage.IsNone() || s.CurrentStage.IsWait()) {
			return nil, study.ErrRoundExists
		}

		// increase total round count
		s.IncrementTotalRound()

		// create new round
		r := study.NewRound()
		r.SetGuildID(s.GuildID)
		r.SetNumber(s.TotalRound)
		r.SetTitle(params.Title)
		r.SetStage(study.StageRegistrationOpened)

		// set initial members
		for _, id := range params.MemberIDs {
			member := study.NewMember()
			r.SetMember(id, member)
		}

		// store new round
		nr, err := svc.tx.CreateRound(sc, r)
		if err != nil {
			return nil, err
		}

		// update study
		s.SetOngoingRoundID(nr.ID)
		s.SetCurrentStage(study.StageRegistrationOpened)

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

// create new study
func (svc *studyService) NewStudy(ctx context.Context, params *NewStudyParams) (*study.Study, error) {
	defer svc.mtx.Unlock()
	svc.mtx.Lock()

	if params == nil {
		return nil, study.ErrNilParams
	}

	txFn := func(sc context.Context) (interface{}, error) {
		// find study of guild
		s, err := svc.tx.FindStudy(sc, params.GuildID)
		if err != nil {
			return nil, err
		}

		// if study exists, return error
		if s != nil {
			return nil, study.ErrStudyExists
		}

		// create new study
		ns := study.New()

		ns.SetGuildID(params.GuildID)
		ns.SetManagerID(params.ManagerID)
		ns.SetNoticeChannelID(params.NoticeChannelID)
		ns.SetReflectionChannelID(params.ReflectionChannelID)

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

// update study and round
func (svc *studyService) UpdateRound(ctx context.Context, params *UpdateParams, update UpdateFunc, validators ...UpdateValidator) (*study.Study, error) {
	defer svc.mtx.Unlock()
	svc.mtx.Lock()

	if params == nil {
		return nil, study.ErrNilParams
	}

	if update == nil {
		return nil, study.ErrNilFunc
	}

	txFn := func(sc context.Context) (interface{}, error) {
		s, err := svc.tx.FindStudy(sc, params.GuildID)
		if err != nil {
			return nil, err
		}

		if s == nil {
			return nil, study.ErrStudyNotFound
		}

		if s.OngoingRoundID == "" {
			return nil, study.ErrRoundNotFound
		}

		r, err := svc.tx.FindRound(sc, s.OngoingRoundID)
		if err != nil {
			return nil, err
		}

		if r == nil {
			return nil, study.ErrRoundNotFound
		}

		// validate
		for _, v := range validators {
			if err := v(s, r, params); err != nil {
				return nil, err
			}
		}

		// update study and round
		update(s, r, params)

		// update study
		s, err = svc.tx.UpdateStudy(sc, *s)
		if err != nil {
			return nil, err
		}

		// update round
		_, err = svc.tx.UpdateRound(sc, *r)
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

// update study
func (svc *studyService) UpdateStudy(ctx context.Context, params *UpdateParams, update UpdateFunc, validators ...UpdateValidator) (*study.Study, error) {
	defer svc.mtx.Unlock()
	svc.mtx.Lock()

	if params == nil {
		return nil, study.ErrNilParams
	}

	if update == nil {
		return nil, study.ErrNilFunc
	}

	txFn := func(sc context.Context) (interface{}, error) {
		s, err := svc.tx.FindStudy(sc, params.GuildID)
		if err != nil {
			return nil, err
		}

		if s == nil {
			return nil, study.ErrStudyNotFound
		}

		// validate
		for _, v := range validators {
			if err := v(s, nil, params); err != nil {
				return nil, err
			}
		}

		// update study
		update(s, nil, params)

		// update study
		s, err = svc.tx.UpdateStudy(sc, *s)
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

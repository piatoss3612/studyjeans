package study

import (
	"context"
	"errors"
	"fmt"
	"sync"

	models "github.com/piatoss3612/presentation-helper-bot/internal/models/study"
	store "github.com/piatoss3612/presentation-helper-bot/internal/store/study"
)

type Service interface {
	GetStudy(ctx context.Context, guildID string) (*models.Study, error)
	GetOngoingRound(ctx context.Context, guildID string) (*models.Round, error)
	GetRounds(ctx context.Context, guildID string) ([]*models.Round, error)

	SetNoticeChannelID(ctx context.Context, guildID, channelID string) error
	SetMemberRegistration(ctx context.Context, guildID, memberID, name, subject string, register bool) error
	SetMemberContent(ctx context.Context, guildID, memberID, contentURL string) error
	SetSpeakerAttended(ctx context.Context, guildID, memberID string, attended bool) error
	SetStudyContent(ctx context.Context, guildID, content string) error
	SetReviewer(ctx context.Context, guildID, reviewerID, revieweeID string) error

	NewStudyRound(ctx context.Context, guildID, title string, memberIDs []string) (*models.Study, error)
	MoveStage(ctx context.Context, guildID string, stage models.Stage) (*models.Study, error)
	CloseStudyRound(ctx context.Context, guildID string) (*models.Study, error)
}

type serviceImpl struct {
	tx store.Tx

	mtx *sync.Mutex
}

// create new service
func NewService(ctx context.Context, tx store.Tx, guildID, managerID, noticeChID string) (Service, error) {
	svc := &serviceImpl{
		tx:  tx,
		mtx: &sync.Mutex{},
	}
	return svc.setup(ctx, guildID, managerID, noticeChID)
}

// setup service
func (svc *serviceImpl) setup(ctx context.Context, guildID, managerID, noticeChID string) (*serviceImpl, error) {
	svc.mtx.Lock()
	defer svc.mtx.Unlock()

	// transaction for setup
	txFn := func(sc context.Context) (interface{}, error) {
		// find study of guild
		s, err := svc.tx.FindStudy(sc, guildID)
		if err != nil {
			return nil, err
		}

		// if there is no management, create one
		if s == nil {
			ns := models.New()

			ns.SetGuildID(guildID)
			ns.SetManagerID(managerID)
			ns.SetNoticeChannelID(noticeChID)

			_, err := svc.tx.CreateStudy(ctx, ns)
			return nil, err
		}
		return nil, nil
	}

	// execute transaction
	_, err := svc.tx.ExecTx(ctx, txFn)
	if err != nil {
		return nil, err
	}

	// return serviceImpl
	return svc, nil
}

// initialize new study round
func (svc *serviceImpl) NewStudyRound(ctx context.Context, guildID, title string, memberIDs []string) (*models.Study, error) {
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
			return nil, ErrStudyNotFound
		}

		// check if there is no ongoing round
		if !(s.CurrentStage.IsNone() || s.CurrentStage.IsWait()) {
			return nil, ErrStudyExists
		}

		// increment total round
		s.IncrementTotalRound()

		// create new round
		r := models.NewRound()
		r.SetGuildID(s.GuildID)
		r.SetNumber(s.TotalRound)
		r.SetTitle(title)

		// set initial members
		for _, id := range memberIDs {
			member := models.NewMember()
			r.SetMember(id, member)
		}

		// store new round
		ur, err := svc.tx.CreateRound(sc, r)
		if err != nil {
			return nil, err
		}

		s.SetOngoingRoundID(ur.ID)
		s.SetCurrentStage(models.StageRegistrationOpened)

		// update study
		return svc.tx.UpdateStudy(sc, *s)
	}

	// execute transaction
	s, err := svc.tx.ExecTx(ctx, txFn)
	if err != nil {
		return nil, err
	}

	// return updated study
	return s.(*models.Study), nil
}

// move study to next stage
func (svc *serviceImpl) MoveStage(ctx context.Context, guildID string, stage models.Stage) (*models.Study, error) {
	defer svc.mtx.Unlock()
	svc.mtx.Lock()

	txFn := func(sc context.Context) (interface{}, error) {
		// find study
		s, err := svc.tx.FindStudy(sc, guildID)
		if err != nil {
			return nil, err
		}

		// if there is no study, return error
		if s == nil || s.CurrentStage.IsNone() || s.CurrentStage.IsWait() {
			return nil, ErrStudyNotFound
		}

		// check if stage is valid
		if !s.CurrentStage.CanMoveTo(stage) {
			return nil, errors.Join(ErrInvalidStage, errors.New(fmt.Sprintf("스터디 라운드 %s 종료가 불가능한 단계입니다.", s.CurrentStage.String())))
		}

		// move to next stage
		s.SetCurrentStage(s.CurrentStage.Next())

		// update study
		return svc.tx.UpdateStudy(sc, *s)
	}

	// execute transaction
	s, err := svc.tx.ExecTx(ctx, txFn)
	if err != nil {
		return nil, err
	}

	// return updated study
	return s.(*models.Study), nil
}

// close study round
func (svc *serviceImpl) CloseStudyRound(ctx context.Context, guildID string) (*models.Study, error) {
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
			return nil, ErrStudyNotFound
		}

		// check if review is finished
		if s.CurrentStage != models.StageReviewClosed {
			return nil, errors.Join(ErrInvalidStage, errors.New(fmt.Sprintf("스터디 라운드 %s 종료가 불가능한 단계입니다.", s.CurrentStage.String())))
		}

		// update management
		s.SetOngoingRoundID("")
		s.SetCurrentStage(models.StageWait)

		// update study
		return svc.tx.UpdateStudy(sc, *s)
	}

	// execute transaction
	s, err := svc.tx.ExecTx(ctx, txFn)
	if err != nil {
		return nil, err
	}
	return s.(*models.Study), nil
}

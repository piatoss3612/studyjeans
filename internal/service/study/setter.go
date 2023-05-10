package study

import (
	"context"
	"errors"

	models "github.com/piatoss3612/presentation-helper-bot/internal/models/study"
)

// set notice channel id
func (svc *serviceImpl) SetNoticeChannelID(ctx context.Context, guildID, channelID string) error {
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

		s.SetNoticeChannelID(channelID)

		// update study
		return svc.tx.UpdateStudy(sc, *s)
	}

	// execute transaction
	_, err := svc.tx.ExecTx(ctx, txFn)
	return err
}

// set member registration
func (svc *serviceImpl) SetMemberRegistration(ctx context.Context, guildID, memberID, name, subject string, register bool) error {
	defer svc.mtx.Unlock()
	svc.mtx.Lock()

	// transaction for changing member registration
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

		if s.CurrentStage != models.StageRegistrationOpened {
			return nil, errors.Join(ErrInvalidStage, errors.New("발표자 등록 및 등록 해지가 불가능한 단계입니다."))
		}

		// if there is no ongoing round, return error
		if s.OngoingRoundID == "" {
			return nil, ErrRoundNotFound
		}

		// find ongoing round
		r, err := svc.tx.FindRound(sc, s.OngoingRoundID)
		if err != nil {
			return nil, err
		}

		// if there is no ongoing round, return error
		if r == nil {
			return nil, ErrRoundNotFound
		}

		// check if member is initialized
		member, ok := r.GetMember(memberID)
		if !ok {
			member = models.NewMember()
		}

		if register {
			// check if member is already registered
			if member.Registered {
				return nil, ErrAlreadyRegistered
			}
			member.SetName(name)
			member.SetSubject(subject)
		} else {
			// check if member is not registered
			if !member.Registered {
				return nil, ErrNotRegistered
			}
			member.SetName("")
			member.SetSubject("")
		}

		member.SetRegistered(register)

		// set updated member to study
		r.SetMember(memberID, member)

		// update round
		return svc.tx.UpdateRound(sc, *r)
	}

	_, err := svc.tx.ExecTx(ctx, txFn)
	return err
}

// set member content
func (svc *serviceImpl) SetMemberContent(ctx context.Context, guildID, memberID, contentURL string) error {
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

		// check if study is in submission stage
		if s.CurrentStage != models.StageSubmissionOpened {
			return nil, errors.Join(ErrInvalidStage, errors.New("발표 자료 제출이 불가능합니다."))
		}

		// if there is no ongoing round, return error
		if s.OngoingRoundID == "" {
			return nil, ErrRoundNotFound
		}

		// find ongoing round
		r, err := svc.tx.FindRound(sc, s.OngoingRoundID)
		if err != nil {
			return nil, err
		}

		// if there is no ongoing round, return error
		if r == nil {
			return nil, ErrRoundNotFound
		}

		// check if member is initialized
		member, ok := r.GetMember(memberID)
		if !ok {
			return nil, ErrMemberNotFound
		}

		// check if member is registered
		if !member.Registered {
			return nil, ErrNotRegistered
		}

		// set content
		member.SetContentURL(contentURL)

		// set updated member to round
		r.SetMember(memberID, member)

		// update round
		return svc.tx.UpdateRound(sc, *r)
	}

	// execute transaction
	_, err := svc.tx.ExecTx(ctx, txFn)
	return err
}

// set speaker attended
func (svc *serviceImpl) SetSpeakerAttended(ctx context.Context, guildID, memberID string, attended bool) error {
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

		// check if presentation is started
		if s.CurrentStage < models.StagePresentationStarted {
			return nil, errors.Join(ErrInvalidStage, errors.New("발표자 출석 확인이 불가능합니다."))
		}

		// if there is no ongoing round, return error
		if s.OngoingRoundID == "" {
			return nil, ErrRoundNotFound
		}

		// find ongoing round
		r, err := svc.tx.FindRound(sc, s.OngoingRoundID)
		if err != nil {
			return nil, err
		}

		// if there is no ongoing round, return error
		if r == nil {
			return nil, ErrRoundNotFound
		}

		// check if member is initialized
		member, ok := r.GetMember(memberID)
		if !ok {
			return nil, ErrMemberNotFound
		}

		// check if member is registered
		if !member.Registered {
			return nil, ErrNotRegistered
		}

		// set attended
		member.SetAttended(attended)

		// set updated member to study
		r.SetMember(memberID, member)

		// update round
		return svc.tx.UpdateRound(sc, *r)
	}

	// execute transaction
	_, err := svc.tx.ExecTx(ctx, txFn)
	return err
}

// set study content
func (svc *serviceImpl) SetStudyContent(ctx context.Context, guildID, content string) error {
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

		// check if presentation is finished
		if s.CurrentStage < models.StagePresentationFinished {
			return nil, errors.Join(ErrInvalidStage, errors.New("스터디 자료 링크 등록이 불가능합니다."))
		}

		// if there is no ongoing round, return error
		if s.OngoingRoundID == "" {
			return nil, ErrRoundNotFound
		}

		// find ongoing round
		r, err := svc.tx.FindRound(sc, s.OngoingRoundID)
		if err != nil {
			return nil, err
		}

		// if there is no ongoing study, return error
		if s == nil {
			return nil, ErrStudyNotFound
		}

		// set content
		r.SetContentURL(content)

		// update round
		return svc.tx.UpdateRound(sc, *r)
	}

	// execute transaction
	_, err := svc.tx.ExecTx(ctx, txFn)
	return err
}

// set reviewer
func (svc *serviceImpl) SetReviewer(ctx context.Context, guildID, reviewerID, revieweeID string) error {
	defer svc.mtx.Unlock()
	svc.mtx.Lock()

	if reviewerID == revieweeID {
		return ErrReviewByYourself
	}

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

		// check if review is ongoing
		if s.CurrentStage < models.StageReviewOpened {
			return nil, errors.Join(ErrInvalidStage, errors.New("리뷰어 지정이 불가능합니다."))
		}

		// if there is no ongoing round, return error
		if s.OngoingRoundID == "" {
			return nil, ErrRoundNotFound
		}

		// find ongoing round
		r, err := svc.tx.FindRound(sc, s.OngoingRoundID)
		if err != nil {
			return nil, err
		}

		// if there is no ongoing round, return error
		if r == nil {
			return nil, ErrRoundNotFound
		}

		// check if reviewer is member of ongoing study
		_, ok := r.GetMember(reviewerID)
		if !ok {
			return nil, errors.Join(ErrMemberNotFound, errors.New("스터디에 참여한 사용자만 리뷰 참여가 가능합니다."))
		}

		// check if reviewee is member of ongoing study
		reviewee, ok := r.GetMember(revieweeID)
		if !ok {
			return nil, errors.Join(ErrMemberNotFound, errors.New("리뷰 대상자는 스터디에 참여한 사용자여야 합니다."))
		}

		// check if reviewee is registered and attended presentation
		if !reviewee.Registered || !reviewee.Attended {
			return nil, errors.New("리뷰 대상자는 발표에 참여한 사용자여야 합니다.")
		}

		// check if reviewer already reviewed
		if reviewee.IsReviewer(reviewerID) {
			return nil, errors.New("이미 리뷰를 작성하였습니다.")
		}

		// set reviewer
		reviewee.SetReviewer(reviewerID)

		// set updated member to study
		r.SetMember(revieweeID, reviewee)

		// update round
		return svc.tx.UpdateRound(sc, *r)
	}

	// execute transaction
	_, err := svc.tx.ExecTx(ctx, txFn)
	return err
}

package service

import (
	"errors"
	"fmt"

	"github.com/piatoss3612/presentation-helper-bot/internal/study"
)

func MoveStage(s *study.Study, r *study.Round, params *UpdateParams) error {
	if s.CurrentStage.IsNone() || s.CurrentStage.IsWait() {
		return study.ErrStudyNotFound
	}

	next := s.CurrentStage.Next()

	if !s.CurrentStage.CanMoveTo(next) {
		return errors.Join(study.ErrInvalidStage, fmt.Errorf("스터디 라운드 '%s'의 종료가 불가능한 단계입니다", s.CurrentStage.String()))
	}

	s.SetCurrentStage(next)
	r.SetStage(next)

	return nil
}

func SetManagerID(s *study.Study, r *study.Round, params *UpdateParams) error {
	if params.ManagerID == "" {
		return errors.Join(study.ErrInvalidUpdateParams, fmt.Errorf("매니저 ID가 없습니다"))
	}

	s.SetManagerID(params.ManagerID)

	return nil
}

func SetNoticeChannelID(s *study.Study, r *study.Round, params *UpdateParams) error {
	if params.ChannelID == "" {
		return errors.Join(study.ErrInvalidUpdateParams, fmt.Errorf("공지 채널 ID가 없습니다"))
	}

	s.SetNoticeChannelID(params.ChannelID)

	return nil
}

func SetReflectionChannelID(s *study.Study, r *study.Round, params *UpdateParams) error {
	if params.ChannelID == "" {
		return errors.Join(study.ErrInvalidUpdateParams, fmt.Errorf("회고 채널 ID가 없습니다"))
	}

	s.SetReflectionChannelID(params.ChannelID)

	return nil
}

func RegisterMemberAsSpeaker(s *study.Study, r *study.Round, params *UpdateParams) error {
	if params.MemberID == "" {
		return errors.Join(study.ErrInvalidUpdateParams, fmt.Errorf("등록할 사용자 ID가 없습니다"))
	}

	if params.MemberName == "" || params.Subject == "" {
		return errors.Join(study.ErrInvalidUpdateParams, fmt.Errorf("발표자 이름 또는 발표 주제가 없습니다"))
	}

	if !s.CurrentStage.IsRegistrationOpened() {
		return errors.Join(study.ErrInvalidStage, fmt.Errorf("발표자 등록이 불가능한 단계입니다"))
	}

	member, ok := r.GetMember(params.MemberID)
	if !ok {
		member = study.NewMember()
	}

	if member.IsRegistered() {
		return study.ErrAlreadyRegistered
	}

	member.SetName(params.MemberName)
	member.SetSubject(params.Subject)
	member.SetRegistered(true)

	r.SetMember(params.MemberID, member)

	return nil
}

func UnregisterSpeaker(s *study.Study, r *study.Round, params *UpdateParams) error {
	if params.MemberID == "" {
		return errors.Join(study.ErrInvalidUpdateParams, fmt.Errorf("등록 해제할 사용자 ID가 없습니다"))
	}

	if !s.CurrentStage.IsRegistrationOpened() {
		return errors.Join(study.ErrInvalidStage, fmt.Errorf("발표자 등록 해제가 불가능한 단계입니다"))
	}

	member, ok := r.GetMember(params.MemberID)
	if !ok {
		return study.ErrMemberNotFound
	}

	if !member.IsRegistered() {
		return study.ErrAlreadyUnregistered
	}

	member.SetRegistered(false)

	r.SetMember(params.MemberID, member)

	return nil
}

func SetMemberContent(s *study.Study, r *study.Round, params *UpdateParams) error {
	if !s.CurrentStage.IsSubmissionOpened() {
		return errors.Join(study.ErrInvalidStage, fmt.Errorf("발표자료 제출이 불가능한 단계입니다"))
	}

	member, ok := r.GetMember(params.MemberID)
	if !ok {
		return study.ErrMemberNotFound
	}

	if !member.IsRegistered() {
		return study.ErrNotRegistered
	}

	member.SetContentURL(params.ContentURL)

	r.SetMember(params.MemberID, member)

	return nil
}

func CheckSpeakerAttendance(s *study.Study, r *study.Round, params *UpdateParams) error {
	if params.MemberID == "" {
		return errors.Join(study.ErrInvalidUpdateParams, fmt.Errorf("출석체크할 사용자 ID가 없습니다"))
	}

	if s.CurrentStage < study.StagePresentationStarted {
		return errors.Join(study.ErrInvalidStage, fmt.Errorf("발표자 출석체크가 불가능한 단계입니다"))
	}

	member, ok := r.GetMember(params.MemberID)
	if !ok {
		return study.ErrMemberNotFound
	}

	if !member.IsRegistered() {
		return study.ErrNotRegistered
	}

	member.SetAttended(true)

	r.SetMember(params.MemberID, member)

	return nil
}

func SetRoundRecordedContent(s *study.Study, r *study.Round, params *UpdateParams) error {
	if params.ContentURL == "" {
		return errors.Join(study.ErrInvalidUpdateParams, fmt.Errorf("발표 녹화본 URL이 없습니다"))
	}

	if s.CurrentStage < study.StagePresentationFinished {
		return errors.Join(study.ErrInvalidStage, fmt.Errorf("발표 녹화본 제출이 불가능한 단계입니다"))
	}

	r.SetContentURL(params.ContentURL)

	return nil
}

func SetReviewer(s *study.Study, r *study.Round, params *UpdateParams) error {
	if params.ReviewerID == "" || params.RevieweeID == "" {
		return errors.Join(study.ErrInvalidUpdateParams, fmt.Errorf("리뷰어 또는 리뷰 대상자 ID가 없습니다"))
	}

	if !s.CurrentStage.IsReviewOpened() {
		return errors.Join(study.ErrInvalidStage, fmt.Errorf("리뷰어 지정이 불가능한 단계입니다"))
	}

	_, ok := r.GetMember(params.ReviewerID)
	if !ok {
		return errors.Join(study.ErrMemberNotFound, errors.New("스터디에 참여한 사용자만 리뷰 참여가 가능합니다"))
	}

	reviewee, ok := r.GetMember(params.RevieweeID)
	if !ok {
		return errors.Join(study.ErrMemberNotFound, errors.New("리뷰 대상자는 발표에 참여한 사용자여야 합니다"))
	}

	if !reviewee.IsRegistered() {
		return study.ErrNotRegistered
	}

	if !reviewee.IsAttended() {
		return study.ErrNotAttended
	}

	if reviewee.IsReviewer(params.ReviewerID) {
		return study.ErrAlreadySentReview
	}

	reviewee.SetReviewer(params.ReviewerID)

	r.SetMember(params.RevieweeID, reviewee)

	return nil
}

func SetSentReflection(s *study.Study, r *study.Round, params *UpdateParams) error {
	if params.MemberID == "" {
		return errors.Join(study.ErrInvalidUpdateParams, fmt.Errorf("회고를 작성할 사용자 ID가 없습니다"))
	}

	if s.CurrentStage < study.StagePresentationFinished {
		return errors.Join(study.ErrInvalidStage, fmt.Errorf("회고 작성이 불가능한 단계입니다"))
	}

	member, ok := r.GetMember(params.MemberID)
	if !ok {
		return study.ErrMemberNotFound
	}

	if !member.IsRegistered() {
		return study.ErrNotRegistered
	}

	if !member.IsAttended() {
		return study.ErrNotAttended
	}

	if member.HasSentReflection() {
		return study.ErrAlreadySentReflection
	}

	member.SetSentReflection(true)

	r.SetMember(params.MemberID, member)

	return nil
}

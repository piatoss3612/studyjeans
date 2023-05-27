package service

import (
	"errors"
	"fmt"

	"github.com/piatoss3612/presentation-helper-bot/internal/study"
)

func ValidateToCheckManager(s *study.Study, r *study.Round, params *UpdateParams) error {
	if !s.IsManager(params.ManagerID) {
		return study.ErrNotManager
	}
	return nil
}

func ValidateToCheckOngoingRound(s *study.Study, r *study.Round, params *UpdateParams) error {
	if s.CurrentStage.IsNone() || s.CurrentStage.IsWait() {
		return study.ErrRoundNotFound
	}
	return nil
}

func ValidateToRegister(s *study.Study, r *study.Round, params *UpdateParams) error {
	if params.MemberID == "" {
		return errors.Join(study.ErrInvalidUpdateParams, fmt.Errorf("등록할 사용자 ID가 없습니다"))
	}

	if !s.CurrentStage.IsRegistrationOpened() {
		return errors.Join(study.ErrInvalidStage, fmt.Errorf("발표자 등록이 불가능한 단계입니다"))
	}

	member, ok := r.GetMember(params.MemberID)
	if ok && member.IsRegistered() {
		return study.ErrAlreadyRegistered
	}

	return nil
}

func ValidateToUnregister(s *study.Study, r *study.Round, params *UpdateParams) error {
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

	return nil
}

func ValidateToSubmitMemberContent(s *study.Study, r *study.Round, params *UpdateParams) error {
	if params.MemberID == "" {
		return errors.Join(study.ErrInvalidUpdateParams, fmt.Errorf("발표자료를 제출할 사용자 ID가 없습니다"))
	}

	if !s.CurrentStage.IsSubmissionOpened() {
		return errors.Join(study.ErrInvalidStage, fmt.Errorf("발표자료 제출이 불가능한 단계입니다"))
	}

	member, ok := r.GetMember(params.MemberID)
	if !ok {
		return study.ErrMemberNotFound
	}

	if !member.IsRegistered() {
		return study.ErrMemberNotRegistered
	}

	return nil
}

func ValidateToCheckAttendance(s *study.Study, r *study.Round, params *UpdateParams) error {
	if params.MemberID == "" {
		return errors.Join(study.ErrInvalidUpdateParams, fmt.Errorf("발표 참여 여부를 확인할 사용자 ID가 없습니다"))
	}

	if s.CurrentStage < study.StagePresentationStarted {
		return errors.Join(study.ErrInvalidStage, fmt.Errorf("발표자 출석체크가 불가능한 단계입니다"))
	}

	member, ok := r.GetMember(params.MemberID)
	if !ok {
		return study.ErrMemberNotFound
	}

	if !member.IsRegistered() {
		return study.ErrMemberNotRegistered
	}

	return nil
}

func ValidateToSubmitRoundContent(s *study.Study, r *study.Round, params *UpdateParams) error {
	if params.ContentURL == "" {
		return errors.Join(study.ErrInvalidUpdateParams, fmt.Errorf("발표 녹화본 URL이 없습니다"))
	}

	if s.CurrentStage < study.StagePresentationFinished {
		return errors.Join(study.ErrInvalidStage, fmt.Errorf("발표 녹화본 제출이 불가능한 단계입니다"))
	}

	return nil
}

func ValidateToSetReviewer(s *study.Study, r *study.Round, params *UpdateParams) error {
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
		return study.ErrMemberNotRegistered
	}

	if !reviewee.IsAttended() {
		return study.ErrMemberNotAttended
	}

	if reviewee.IsReviewer(params.ReviewerID) {
		return study.ErrAlreadySentReview
	}

	return nil
}

func ValidateToSetSendReflection(s *study.Study, r *study.Round, params *UpdateParams) error {
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
		return study.ErrMemberNotRegistered
	}

	if !member.IsAttended() {
		return study.ErrMemberNotAttended
	}

	if member.HasSentReflection() {
		return study.ErrAlreadySentReflection
	}

	return nil
}

package study

import (
	"errors"
	"sync"
)

type Service struct {
	Management   *Management
	OnGoingStudy *Study
	// Store

	mtx *sync.RWMutex
}

func NewService(guildID string) (*Service, error) {
	svc := &Service{
		mtx: &sync.RWMutex{},
	}
	return svc.setup(guildID)
}

func (s *Service) setup(guildID string) (*Service, error) {
	// TODO: get study manage from repository
	return s, nil
}

func (s *Service) GetNoticeChannelID() string {
	defer s.mtx.RUnlock()
	s.mtx.RLock()

	return s.Management.NoticeChannelID
}

func (s *Service) SetNoticeChannelID(channelID string) {
	defer s.mtx.Unlock()
	s.mtx.Lock()

	s.Management.SetNoticeChannelID(channelID)
}

func (s *Service) CreateStudy(proposerID, title string, memberIDs []string) error {
	defer s.mtx.Unlock()
	s.mtx.Lock()

	m := *s.Management

	// check if proposer is manager
	if !m.IsManager(proposerID) {
		return errors.New("스터디 관리자가 아닙니다.")
	}

	// check if there is no on going study
	if !(m.CurrentStudyStage.IsNone() || m.CurrentStudyStage.IsWait()) {
		return errors.New("이미 진행중인 스터디가 있습니다.")
	}

	// create study
	study := New()

	study.SetGuildID(m.GuildID)
	study.SetTitle(title)

	// set initial members
	for _, id := range memberIDs {
		member := NewMember()
		study.SetMember(id, member)
	}

	// TODO: save study to repository

	// move to next stage
	m.SetOnGoingStudyID(study.ID)
	m.SetCurrentStudyStage(StudyStageRegistrationStarted)

	// TODO: save manage to repository

	// TODO: commit transaction

	// set study and manage
	s.Management = &m
	s.OnGoingStudy = study

	return nil
}

func (s *Service) ChangeMemberRegistration(guildID, memberID, name, subject string, registered bool) error {
	defer s.mtx.Unlock()
	s.mtx.Lock()

	m := s.Management

	// check if study is in registration stage
	if !m.CurrentStudyStage.IsRegistrationOngoing() {
		return errors.New("발표자 등록 및 등록 해지가 불가능한 상태입니다.")
	}

	study := *s.OnGoingStudy

	// check if presentor belongs to study
	if study.GuildID != guildID {
		return errors.New("해당 디스코드 서버에서 진행중인 스터디가 아닙니다.")
	}

	// check if presentor is initialized
	member, ok := study.GetMember(memberID)
	if !ok {
		member = NewMember()
	}

	// change member's registered state
	member.SetName(name)
	member.SetSubject(subject)
	member.SetRegistered(registered)

	study.SetMember(memberID, member)

	// TODO: save study to repository

	s.OnGoingStudy = &study

	return nil
}

func (s *Service) FinishRegistration(proposerID string) error {
	defer s.mtx.Unlock()
	s.mtx.Lock()

	m := *s.Management

	// check if proposer is manager
	if !m.IsManager(proposerID) {
		return errors.New("스터디 관리자가 아닙니다.")
	}

	// check if study is in registration stage
	if !m.CurrentStudyStage.IsRegistrationOngoing() {
		return errors.New("발표자 등록 완료가 불가능한 상태입니다.")
	}

	m.SetCurrentStudyStage(StudyStageRegistrationFinished)

	// TODO: save manage to repository

	s.Management = &m

	return nil
}

func (s *Service) StartSubmission(proposerID string) error {
	defer s.mtx.Unlock()
	s.mtx.Lock()

	m := *s.Management

	// check if proposer is manager
	if !m.IsManager(proposerID) {
		return errors.New("스터디 관리자가 아닙니다.")
	}

	// check if study can accept content submission
	if m.CurrentStudyStage.IsRegistrationFinished() {
		return errors.New("발표 자료 제출 단계 시작이 불가능한 상태입니다.")
	}

	m.SetCurrentStudyStage(StudyStageSubmissionStarted)

	// TODO: save manage to repository

	s.Management = &m

	return nil
}

func (s *Service) SubmitContent(guildID, memberID, contentURL string) error {
	defer s.mtx.Unlock()
	s.mtx.Lock()

	m := s.Management

	// check if study can accept content submission
	if !m.CurrentStudyStage.IsSubmissionOngoing() {
		return errors.New("발표 자료 제출이 불가능한 상태입니다.")
	}

	study := *s.OnGoingStudy

	// check if presentor belongs to study
	if study.GuildID != guildID {
		return errors.New("해당 디스코드 서버에서 진행중인 스터디가 아닙니다.")
	}

	// check if presentor is initialized
	member, ok := study.GetMember(memberID)
	if !ok {
		return errors.New("활성화된 스터디에 등록되지 않은 사용자입니다.")
	}

	// check if presentor is registered
	if !member.Registered {
		return errors.New("발표자로 등록되지 않은 사용자입니다.")
	}

	// set content
	member.SetContentURL(contentURL)
	study.SetMember(memberID, member)

	// TODO: save study to repository

	s.OnGoingStudy = &study

	return nil
}

func (s *Service) FinishSubmission(proposerID string) error {
	defer s.mtx.Unlock()
	s.mtx.Lock()

	m := *s.Management

	// check if proposer is manager
	if !m.IsManager(proposerID) {
		return errors.New("스터디 관리자가 아닙니다.")
	}

	// check if study can accept content submission
	if !m.CurrentStudyStage.IsSubmissionOngoing() {
		return errors.New("발표 자료 제출 단계 종료가 불가능한 상태입니다.")
	}

	m.SetCurrentStudyStage(StudyStageSubmissionFinished)

	// TODO: save manage to repository

	s.Management = &m

	return nil
}

func (s *Service) StartPresentation(proposerID string) error {
	defer s.mtx.Unlock()
	s.mtx.Lock()

	m := *s.Management

	// check if proposer is manager
	if !m.IsManager(proposerID) {
		return errors.New("스터디 관리자가 아닙니다.")
	}

	// check if content submission is finished
	if !m.CurrentStudyStage.IsSubmissionFinished() {
		return errors.New("발표 단계 시작이 불가능한 상태입니다.")
	}

	m.SetCurrentStudyStage(StudyStagePresentationStarted)

	// TODO: save study manage to repository

	s.Management = &m

	return nil
}

func (s *Service) ChangePresentationParticipated(proposerID, memberID string, participated bool) error {
	defer s.mtx.Unlock()
	s.mtx.Lock()

	m := *s.Management

	// check if proposer is manager
	if !m.IsManager(proposerID) {
		return errors.New("스터디 관리자가 아닙니다.")
	}

	// check if presentation is ongoing
	if !m.CurrentStudyStage.IsPresentationOngoing() {
		return errors.New("발표 완료 상태 전환이 불가능한 상태입니다.")
	}

	study := *s.OnGoingStudy

	// check if presentor is initialized
	member, ok := study.GetMember(memberID)
	if !ok {
		return errors.New("활성화된 스터디에 등록되지 않은 사용자입니다.")
	}

	// check if presentor is registered
	if !member.Registered {
		return errors.New("발표자로 등록되지 않은 사용자입니다.")
	}

	// set complete state
	member.SetParticipated(participated)
	study.SetMember(memberID, member)

	// TODO: save study to repository

	s.OnGoingStudy = &study

	return nil
}

func (s *Service) FinishPresentation(proposerID string) error {
	defer s.mtx.Unlock()
	s.mtx.Lock()

	m := *s.Management

	// check if proposer is manager
	if !m.IsManager(proposerID) {
		return errors.New("스터디 관리자가 아닙니다.")
	}

	// check if presentation is ongoing
	if !m.CurrentStudyStage.IsPresentationOngoing() {
		return errors.New("발표 단계 종료가 불가능한 상태입니다.")
	}

	m.SetCurrentStudyStage(StudyStagePresentationFinished)

	// TODO: save study manage to repository

	s.Management = &m

	return nil
}

func (s *Service) StartReview(proposerID string) error {
	defer s.mtx.Unlock()
	s.mtx.Lock()

	m := *s.Management

	// check if proposer is manager
	if !m.IsManager(proposerID) {
		return errors.New("스터디 관리자가 아닙니다.")
	}

	// check if presentation is finished
	if !m.CurrentStudyStage.IsPresentationFinished() {
		return errors.New("리뷰 단계 시작이 불가능한 상태입니다.")
	}

	m.SetCurrentStudyStage(StudyStageReviewStarted)

	// TODO: save study manage to repository

	s.Management = &m

	return nil
}

func (s *Service) IsReviewer(guildID, reviewerID, revieweeID string) (bool, error) {
	defer s.mtx.RUnlock()
	s.mtx.RLock()

	m := s.Management

	if !m.CurrentStudyStage.IsReviewOngoing() {
		return false, errors.New("리뷰 단계가 진행중이 아닙니다.")
	}

	study := s.OnGoingStudy

	// check if reviewee belongs to study
	if study.GuildID != guildID {
		return false, errors.New("스터디 서버 정보가 일치하지 않습니다.")
	}

	reviewee, ok := study.GetMember(revieweeID)
	if !ok {
		return false, errors.New("활성화된 스터디에 등록되지 않은 사용자입니다.")
	}

	// check if reviewee is registered
	if !reviewee.Registered {
		return false, errors.New("발표자로 등록되지 않은 사용자입니다.")
	}

	// check if reviewee participated presentation
	if !reviewee.Participated {
		return false, errors.New("발표를 완료하지 않은 사용자입니다.")
	}

	// check if reviewer already reviewed
	return reviewee.IsReviewer(reviewerID), nil
}

func (s *Service) SetReviewer(guildID, reviewerID, revieweeID string) error {
	defer s.mtx.Unlock()
	s.mtx.Lock()

	m := s.Management

	if !m.CurrentStudyStage.IsReviewOngoing() {
		return errors.New("리뷰 단계가 진행중이 아닙니다.")
	}

	study := *s.OnGoingStudy

	// check if reviewee belongs to study
	if study.GuildID != guildID {
		return errors.New("스터디 서버 정보가 일치하지 않습니다.")
	}

	reviewee, ok := study.GetMember(revieweeID)
	if !ok {
		return errors.New("활성화된 스터디에 등록되지 않은 사용자입니다.")
	}

	// check if reviewee is registered
	if !reviewee.Registered {
		return errors.New("발표자로 등록되지 않은 사용자입니다.")
	}

	// check if reviewee participated presentation
	if !reviewee.Participated {
		return errors.New("발표에 참여하지 않은 사용자입니다.")
	}

	// check if reviewer already reviewed
	if reviewee.IsReviewer(reviewerID) {
		return errors.New("이미 리뷰를 완료한 사용자입니다.")
	}

	// set reviewer
	reviewee.SetReviewer(reviewerID)
	study.SetMember(revieweeID, reviewee)

	// TODO: save study to repository

	s.OnGoingStudy = &study

	return nil
}

func (s *Service) FinishReview(proposerID string) error {
	defer s.mtx.Unlock()
	s.mtx.Lock()

	m := *s.Management

	// check if proposer is manager
	if !m.IsManager(proposerID) {
		return errors.New("스터디 관리자가 아닙니다.")
	}

	// check if review is ongoing
	if !m.CurrentStudyStage.IsReviewOngoing() {
		return errors.New("리뷰 단계 종료가 불가능한 상태입니다.")
	}

	m.SetCurrentStudyStage(StudyStageReviewFinished)

	// TODO: save study manage to repository

	s.Management = &m

	return nil
}

func (s *Service) FinishStudy(proposerID string) error {
	defer s.mtx.Unlock()
	s.mtx.Lock()

	m := *s.Management

	// check if proposer is manager
	if !m.IsManager(proposerID) {
		return errors.New("스터디 관리자가 아닙니다.")
	}

	// check if review is finished
	if !m.CurrentStudyStage.IsReviewFinished() {
		return errors.New("스터디 종료가 불가능한 상태입니다.")
	}

	m.SetOnGoingStudyID("")
	m.SetCurrentStudyStage(StudyStageWait)

	// TODO: save study manage to repository

	s.Management = &m
	s.OnGoingStudy = nil

	return nil
}

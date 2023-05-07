package study

import (
	"context"
	"errors"
	"sync"
	"time"
)

type Service interface {
	GetNoticeChannelID() string
	SetNoticeChannelID(ctx context.Context, proposerID, channelID string) error
	GetStudies(ctx context.Context, guildID string) ([]*Study, error)
	CreateStudy(ctx context.Context, proposerID, title string, memberIDs []string) error
	ChangeMemberRegistration(ctx context.Context, guildID, memberID, name, subject string, registered bool) error
	FinishRegistration(ctx context.Context, proposerID string) error
	StartSubmission(ctx context.Context, proposerID string) error
	SubmitContent(ctx context.Context, guildID, memberID, contentURL string) error
	FinishSubmission(ctx context.Context, proposerID string) error
	StartPresentation(ctx context.Context, proposerID string) error
	ChangePresentationAttended(ctx context.Context, proposerID, memberID string, attended bool) error
	FinishPresentation(ctx context.Context, proposerID string) error
	StartReview(ctx context.Context, proposerID string) error
	SetReviewer(ctx context.Context, guildID, reviewerID, revieweeID string) error
	FinishReview(ctx context.Context, proposerID string) error
	FinishStudy(ctx context.Context, proposerID string) error
}

type ServiceImpl struct {
	Management   *Management
	OnGoingStudy *Study
	Tx           Tx

	mtx *sync.RWMutex
}

func NewService(ctx context.Context, tx Tx, guildID, managerID, noticeChID string) (Service, error) {
	svc := &ServiceImpl{
		Tx:  tx,
		mtx: &sync.RWMutex{},
	}
	return svc.setup(ctx, guildID, managerID, noticeChID)
}

func (s *ServiceImpl) setup(ctx context.Context, guildID, managerID, noticeChID string) (*ServiceImpl, error) {
	// transaction for setup
	tx := func(sc context.Context) (interface{}, error) {
		// find management
		m, err := s.Tx.FindManagement(sc, guildID)
		if err != nil {
			return nil, err
		}

		// if there is no management, create one
		if m == nil {
			m = NewManagement()
			m.SetGuildID(guildID)
			m.SetManagerID(managerID)
			m.SetNoticeChannelID(noticeChID)

			id, err := s.Tx.StoreManagement(ctx, *m)
			if err != nil {
				return nil, err
			}

			m.SetID(id)
		}

		// set management
		s.Management = m

		// if there is no ongoing study, return
		if m.OngoingStudyID == "" {
			return nil, nil
		}

		// find ongoing study
		study, err := s.Tx.FindStudy(sc, m.OngoingStudyID)
		if err != nil {
			return nil, err
		}

		// if there is no ongoing study, return error
		if study == nil {
			return nil, errors.New("진행중인 스터디를 찾을 수 없습니다.")
		}

		// set ongoing study
		s.OnGoingStudy = study

		return nil, nil
	}

	// execute transaction
	_, err := s.Tx.ExecTx(ctx, tx)
	if err != nil {
		return nil, err
	}

	// return ServiceImpl
	return s, nil
}

func (s *ServiceImpl) GetNoticeChannelID() string {
	defer s.mtx.RUnlock()
	s.mtx.RLock()

	return s.Management.NoticeChannelID
}

func (s *ServiceImpl) SetNoticeChannelID(ctx context.Context, proposerID, channelID string) error {
	defer s.mtx.Unlock()
	s.mtx.Lock()

	// check if management is initialized
	if s.Management == nil {
		return errors.New("스터디 관리가 설정되지 않았습니다.")
	}

	m := *s.Management

	// check if proposer is manager
	if !m.IsManager(proposerID) {
		return errors.New("스터디 관리자가 아닙니다.")
	}

	m.SetNoticeChannelID(channelID)

	// update management
	err := s.Tx.UpdateManagement(ctx, m)
	if err != nil {
		return err
	}

	s.Management = &m

	return nil
}

func (s *ServiceImpl) GetStudies(ctx context.Context, guildID string) ([]*Study, error) {
	defer s.mtx.RUnlock()
	s.mtx.RLock()

	// check if management is initialized
	if s.Management == nil {
		return nil, errors.New("스터디 관리가 설정되지 않았습니다.")
	}

	m := *s.Management

	// check if guildID is same
	if m.GuildID != guildID {
		return nil, errors.New("서버 ID가 일치하지 않습니다.")
	}

	// find studies
	return s.Tx.FindStudies(ctx, guildID)
}

func (s *ServiceImpl) CreateStudy(ctx context.Context, proposerID, title string, memberIDs []string) error {
	defer s.mtx.Unlock()
	s.mtx.Lock()

	// check if management is initialized
	if s.Management == nil {
		return errors.New("스터디 관리가 설정되지 않았습니다.")
	}

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

	// move to next stage
	m.SetOngoingStudyID(study.ID)
	m.SetCurrentStudyStage(StudyStageRegistrationStarted)

	// transaction for create study and update management
	tx := func(sc context.Context) (interface{}, error) {
		// store new study
		studyID, err := s.Tx.StoreStudy(sc, *study)
		if err != nil {
			return nil, err
		}

		// set study id to study and management
		study.SetID(studyID)
		m.SetOngoingStudyID(studyID)

		// update management
		err = s.Tx.UpdateManagement(sc, m)
		if err != nil {
			return nil, err
		}

		return nil, nil
	}

	// execute transaction
	_, err := s.Tx.ExecTx(ctx, tx)
	if err != nil {
		return err
	}

	// set study and manage
	s.Management = &m
	s.OnGoingStudy = study

	return nil
}

func (s *ServiceImpl) ChangeMemberRegistration(ctx context.Context, guildID, memberID, name, subject string, registered bool) error {
	defer s.mtx.Unlock()
	s.mtx.Lock()

	// check if management is initialized
	if s.Management == nil {
		return errors.New("스터디 관리가 설정되지 않았습니다.")
	}

	m := s.Management

	// check if study is in registration stage
	if !m.CurrentStudyStage.IsRegistrationOngoing() {
		return errors.New("발표자 등록 및 등록 해지가 불가능한 상태입니다.")
	}

	// check if there is no ongoing study
	if m.OngoingStudyID == "" || s.OnGoingStudy == nil {
		return errors.New("진행중인 스터디가 없습니다.")
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

	// set updated member to study
	study.SetMember(memberID, member)
	study.SetUpdatedAt(time.Now())

	// update study
	err := s.Tx.UpdateStudy(ctx, study)
	if err != nil {
		return err
	}

	s.OnGoingStudy = &study

	return nil
}

func (s *ServiceImpl) FinishRegistration(ctx context.Context, proposerID string) error {
	defer s.mtx.Unlock()
	s.mtx.Lock()

	// check if management is initialized
	if s.Management == nil {
		return errors.New("스터디 관리가 설정되지 않았습니다.")
	}

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
	m.SetUpdatedAt(time.Now())

	err := s.Tx.UpdateManagement(ctx, m)
	if err != nil {
		return err
	}

	s.Management = &m

	return nil
}

func (s *ServiceImpl) StartSubmission(ctx context.Context, proposerID string) error {
	defer s.mtx.Unlock()
	s.mtx.Lock()

	// check if management is initialized
	if s.Management == nil {
		return errors.New("스터디 관리가 설정되지 않았습니다.")
	}

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
	m.SetUpdatedAt(time.Now())

	// update management
	err := s.Tx.UpdateManagement(ctx, m)
	if err != nil {
		return err
	}

	s.Management = &m

	return nil
}

func (s *ServiceImpl) SubmitContent(ctx context.Context, guildID, memberID, contentURL string) error {
	defer s.mtx.Unlock()
	s.mtx.Lock()

	// check if management is initialized
	if s.Management == nil {
		return errors.New("스터디 관리가 설정되지 않았습니다.")
	}

	m := s.Management

	// check if study can accept content submission
	if !m.CurrentStudyStage.IsSubmissionOngoing() {
		return errors.New("발표 자료 제출이 불가능한 상태입니다.")
	}

	// check if there is no ongoing study
	if m.OngoingStudyID == "" || s.OnGoingStudy == nil {
		return errors.New("진행중인 스터디가 없습니다.")
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

	// set updated member to study
	study.SetMember(memberID, member)
	study.SetUpdatedAt(time.Now())

	// update study
	err := s.Tx.UpdateStudy(ctx, study)
	if err != nil {
		return err
	}

	s.OnGoingStudy = &study

	return nil
}

func (s *ServiceImpl) FinishSubmission(ctx context.Context, proposerID string) error {
	defer s.mtx.Unlock()
	s.mtx.Lock()

	// check if management is initialized
	if s.Management == nil {
		return errors.New("스터디 관리가 설정되지 않았습니다.")
	}

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
	m.SetUpdatedAt(time.Now())

	// update management
	err := s.Tx.UpdateManagement(ctx, m)
	if err != nil {
		return err
	}

	s.Management = &m

	return nil
}

func (s *ServiceImpl) StartPresentation(ctx context.Context, proposerID string) error {
	defer s.mtx.Unlock()
	s.mtx.Lock()

	// check if management is initialized
	if s.Management == nil {
		return errors.New("스터디 관리가 설정되지 않았습니다.")
	}

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
	m.SetUpdatedAt(time.Now())

	// update management
	err := s.Tx.UpdateManagement(ctx, m)
	if err != nil {
		return err
	}

	s.Management = &m

	return nil
}

func (s *ServiceImpl) ChangePresentationAttended(ctx context.Context, proposerID, memberID string, attended bool) error {
	defer s.mtx.Unlock()
	s.mtx.Lock()

	// check if management is initialized
	if s.Management == nil {
		return errors.New("스터디 관리가 설정되지 않았습니다.")
	}

	m := *s.Management

	// check if proposer is manager
	if !m.IsManager(proposerID) {
		return errors.New("스터디 관리자가 아닙니다.")
	}

	// check if presentation is ongoing
	if !m.CurrentStudyStage.IsPresentationOngoing() {
		return errors.New("발표 완료 상태 전환이 불가능한 상태입니다.")
	}

	// check if there is no ongoing study
	if m.OngoingStudyID == "" || s.OnGoingStudy == nil {
		return errors.New("진행중인 스터디가 없습니다.")
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
	member.SetAttended(attended)

	// set updated member to study
	study.SetMember(memberID, member)
	study.SetUpdatedAt(time.Now())

	// update study
	err := s.Tx.UpdateStudy(ctx, study)
	if err != nil {
		return err
	}

	s.OnGoingStudy = &study

	return nil
}

func (s *ServiceImpl) FinishPresentation(ctx context.Context, proposerID string) error {
	defer s.mtx.Unlock()
	s.mtx.Lock()

	// check if management is initialized
	if s.Management == nil {
		return errors.New("스터디 관리가 설정되지 않았습니다.")
	}

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
	m.SetUpdatedAt(time.Now())

	// update management
	err := s.Tx.UpdateManagement(ctx, m)
	if err != nil {
		return err
	}

	s.Management = &m

	return nil
}

func (s *ServiceImpl) StartReview(ctx context.Context, proposerID string) error {
	defer s.mtx.Unlock()
	s.mtx.Lock()

	// check if management is initialized
	if s.Management == nil {
		return errors.New("스터디 관리가 설정되지 않았습니다.")
	}

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
	m.SetUpdatedAt(time.Now())

	// update management
	err := s.Tx.UpdateManagement(ctx, m)
	if err != nil {
		return err
	}

	s.Management = &m

	return nil
}

func (s *ServiceImpl) SetReviewer(ctx context.Context, guildID, reviewerID, revieweeID string) error {
	defer s.mtx.Unlock()
	s.mtx.Lock()

	// check if management is initialized
	if s.Management == nil {
		return errors.New("스터디 관리가 설정되지 않았습니다.")
	}

	m := s.Management

	if !m.CurrentStudyStage.IsReviewOngoing() {
		return errors.New("리뷰 단계가 진행중이 아닙니다.")
	}

	// check if there is no ongoing study
	if m.OngoingStudyID == "" || s.OnGoingStudy == nil {
		return errors.New("진행중인 스터디가 없습니다.")
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

	// check if reviewee Attended presentation
	if !reviewee.Attended {
		return errors.New("발표에 참여하지 않은 사용자입니다.")
	}

	// check if reviewer already reviewed
	if reviewee.IsReviewer(reviewerID) {
		return errors.New("이미 리뷰를 완료하였습니다.")
	}

	// set reviewer
	reviewee.SetReviewer(reviewerID)

	// set updated member to study
	study.SetMember(revieweeID, reviewee)
	study.SetUpdatedAt(time.Now())

	// update study
	err := s.Tx.UpdateStudy(ctx, study)
	if err != nil {
		return err
	}

	s.OnGoingStudy = &study

	return nil
}

func (s *ServiceImpl) FinishReview(ctx context.Context, proposerID string) error {
	defer s.mtx.Unlock()
	s.mtx.Lock()

	// check if management is initialized
	if s.Management == nil {
		return errors.New("스터디 관리가 설정되지 않았습니다.")
	}

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
	m.SetUpdatedAt(time.Now())

	// update management
	err := s.Tx.UpdateManagement(ctx, m)
	if err != nil {
		return err
	}

	s.Management = &m

	return nil
}

func (s *ServiceImpl) FinishStudy(ctx context.Context, proposerID string) error {
	defer s.mtx.Unlock()
	s.mtx.Lock()

	// check if management is initialized
	if s.Management == nil {
		return errors.New("스터디 관리가 설정되지 않았습니다.")
	}

	m := *s.Management

	// check if proposer is manager
	if !m.IsManager(proposerID) {
		return errors.New("스터디 관리자가 아닙니다.")
	}

	// check if review is finished
	if !m.CurrentStudyStage.IsReviewFinished() {
		return errors.New("스터디 종료가 불가능한 상태입니다.")
	}

	m.SetOngoingStudyID("")
	m.SetCurrentStudyStage(StudyStageWait)
	m.SetUpdatedAt(time.Now())

	// update management
	err := s.Tx.UpdateManagement(ctx, m)
	if err != nil {
		return err
	}

	s.Management = &m
	s.OnGoingStudy = nil

	return nil
}

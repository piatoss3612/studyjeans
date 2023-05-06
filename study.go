package main

import (
	"errors"
	"sync"
	"time"
)

type StudyStage uint8

const (
	StudyStageNone                 StudyStage = 0
	StudyStageWait                 StudyStage = 1
	StudyStageRegistrationStarted  StudyStage = 2
	StudyStageRegistrationFinished StudyStage = 3
	StudyStageSubmissionStarted    StudyStage = 4
	StudyStageSubmissionFinished   StudyStage = 5
	StudyStagePresentationStarted  StudyStage = 6
	StudyStagePresentationFinished StudyStage = 7
	StudyStageReviewStarted        StudyStage = 8
	StudyStageReviewFinished       StudyStage = 9
)

func (s StudyStage) String() string {
	switch s {
	case StudyStageWait:
		return "다음 회차 대기"
	case StudyStageRegistrationStarted, StudyStageRegistrationFinished:
		return "발표자 등록"
	case StudyStageSubmissionStarted, StudyStageSubmissionFinished:
		return "발표자료 제출"
	case StudyStagePresentationStarted, StudyStagePresentationFinished:
		return "발표"
	case StudyStageReviewStarted, StudyStageReviewFinished:
		return "리뷰 및 피드백"
	default:
		return "몰?루"
	}
}

func (s StudyStage) IsNone() bool {
	return s == StudyStageNone
}

func (s StudyStage) IsWait() bool {
	return s == StudyStageWait
}

func (s StudyStage) IsRegistrationOngoing() bool {
	return s == StudyStageRegistrationStarted
}

func (s StudyStage) IsRegistrationFinished() bool {
	return s == StudyStageRegistrationFinished
}

func (s StudyStage) IsSubmissionOngoing() bool {
	return s == StudyStageSubmissionStarted
}

func (s StudyStage) IsSubmissionFinished() bool {
	return s == StudyStageSubmissionFinished
}

func (s StudyStage) IsPresentationOngoing() bool {
	return s == StudyStagePresentationStarted
}

func (s StudyStage) IsPresentationFinished() bool {
	return s == StudyStagePresentationFinished
}

func (s StudyStage) IsReviewOngoing() bool {
	return s == StudyStageReviewStarted
}

func (s StudyStage) IsReviewFinished() bool {
	return s == StudyStageReviewFinished
}

type StudyManager struct {
	GuildID         string
	NoticeChannelID string

	ManagerID     string
	SubManagerIDs []string

	OnGoingStudyID string
	StudyStage     StudyStage

	mtx *sync.RWMutex
}

func NewStudyManager(guildID string, ManagerID string) *StudyManager {
	return &StudyManager{
		GuildID:         guildID,
		NoticeChannelID: "",
		ManagerID:       ManagerID,
		SubManagerIDs:   []string{},
		OnGoingStudyID:  "",
		StudyStage:      StudyStageNone,
		mtx:             &sync.RWMutex{},
	}
}

func (s *StudyManager) SetNoticeChannelID(channelID string) {
	defer s.mtx.Unlock()
	s.mtx.Lock()

	s.NoticeChannelID = channelID
}

func (s *StudyManager) IsManager(userID string) bool {
	defer s.mtx.RLock()
	s.mtx.RLock()

	return s.ManagerID == userID
}

func (s *StudyManager) AddSubManagerID(userID string) {
	defer s.mtx.Unlock()
	s.mtx.Lock()

	s.SubManagerIDs = append(s.SubManagerIDs, userID)
}

func (s *StudyManager) RemoveSubManagerID(userID string) {
	defer s.mtx.Unlock()
	s.mtx.Lock()

	for i, v := range s.SubManagerIDs {
		if v == userID {
			s.SubManagerIDs = append(s.SubManagerIDs[:i], s.SubManagerIDs[i+1:]...)
			return
		}
	}
}

func (s *StudyManager) IsSubManager(userID string) bool {
	defer s.mtx.RLock()
	s.mtx.RLock()

	for _, v := range s.SubManagerIDs {
		if v == userID {
			return true
		}
	}
	return false
}

func (s *StudyManager) SetOnGoingStudyID(studyID string) {
	defer s.mtx.Unlock()
	s.mtx.Lock()

	s.OnGoingStudyID = studyID
}

func (s *StudyManager) SetStudyStage(state StudyStage) {
	defer s.mtx.Unlock()
	s.mtx.Lock()

	s.StudyStage = state
}

type Member struct {
	Name       string
	Registered bool

	Subject    string
	ContentURL string
	Completed  bool
	Reviewers  map[string]bool

	mtx *sync.RWMutex
}

func NewMember(name string) Member {
	return Member{
		Name:       name,
		Registered: false,
		Subject:    "",
		ContentURL: "",
		Completed:  false,
		Reviewers:  map[string]bool{},
		mtx:        &sync.RWMutex{},
	}
}

func (m *Member) SetRegistered(registered bool) {
	defer m.mtx.Unlock()
	m.mtx.Lock()

	m.Registered = registered
}

func (m *Member) SetSubject(subject string) {
	defer m.mtx.Unlock()
	m.mtx.Lock()

	m.Subject = subject
}

func (m *Member) SetContentURL(contentURL string) {
	defer m.mtx.Unlock()
	m.mtx.Lock()

	m.ContentURL = contentURL
}

func (m *Member) SetCompleted(completed bool) {
	defer m.mtx.Unlock()
	m.mtx.Lock()

	m.Completed = completed
}

func (m *Member) SetReviewer(userID string) {
	defer m.mtx.Unlock()
	m.mtx.Lock()

	m.Reviewers[userID] = true
}

func (m *Member) HasDoneReview(userID string) bool {
	defer m.mtx.RUnlock()
	m.mtx.RLock()

	return m.Reviewers[userID]
}

type Study struct {
	ID      string
	GuildID string

	Title     string
	Members   map[string]Member
	CreatedAt time.Time
	UpdatedAt time.Time

	mtx *sync.RWMutex
}

func NewStudy(guildID, title string) *Study {
	return &Study{
		ID:        "",
		GuildID:   guildID,
		Title:     title,
		Members:   map[string]Member{},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		mtx:       &sync.RWMutex{},
	}
}

func (s *Study) SetID(id string) {
	defer s.mtx.Unlock()
	s.mtx.Lock()

	s.ID = id
}

func (s *Study) SetTitle(title string) {
	defer s.mtx.Unlock()
	s.mtx.Lock()

	s.Title = title
}

func (s *Study) SetMember(memberID string, member Member) {
	defer s.mtx.Unlock()
	s.mtx.Lock()

	s.Members[memberID] = member
}

func (s *Study) RemoveMember(memberID string) {
	defer s.mtx.Unlock()
	s.mtx.Lock()

	delete(s.Members, memberID)
}

func (s *Study) GetMember(memberID string) (Member, bool) {
	defer s.mtx.RUnlock()
	s.mtx.RLock()

	member, ok := s.Members[memberID]
	return member, ok
}

func (s *Study) GetMembers() []Member {
	defer s.mtx.RUnlock()
	s.mtx.RLock()
	members := []Member{}
	for _, v := range s.Members {
		members = append(members, v)
	}
	return members
}

func (s *Study) SetUpdatedAt(updatedAt time.Time) {
	defer s.mtx.Unlock()
	s.mtx.Lock()

	s.UpdatedAt = updatedAt
}

type StudyService struct {
	StudyManager *StudyManager
	OnGoingStudy *Study
	// Repository

	mtx *sync.RWMutex
}

func NewStudyService(guildID string) (*StudyService, error) {
	svc := &StudyService{
		mtx: &sync.RWMutex{},
	}
	return svc.setup(guildID)
}

func (s *StudyService) setup(guildID string) (*StudyService, error) {
	// TODO: get study manager from repository
	return s, nil
}

func (s *StudyService) GetNoticeChannelID() string {
	defer s.mtx.RUnlock()
	s.mtx.RLock()

	return s.StudyManager.NoticeChannelID
}

func (s *StudyService) SetNoticeChannelID(channelID string) {
	defer s.mtx.Unlock()
	s.mtx.Lock()

	s.StudyManager.SetNoticeChannelID(channelID)
}

func (s *StudyService) CreateStudy(proposerID, guildID, title string, memberIDs []string) error {
	defer s.mtx.Unlock()
	s.mtx.Lock()

	manager := *s.StudyManager

	// check if proposer is manager
	if !manager.IsManager(proposerID) {
		return errors.New("스터디 관리자가 아닙니다.")
	}

	// check if there is no on going study
	if !(manager.StudyStage.IsNone() || manager.StudyStage.IsWait()) {
		return errors.New("이미 진행중인 스터디가 있습니다.")
	}

	// create study
	study := NewStudy(guildID, title)

	// set initial members
	for _, id := range memberIDs {
		member := NewMember(id)
		study.SetMember(id, member)
	}

	// TODO: save study to repository

	// move to next stage
	manager.SetOnGoingStudyID(study.ID)
	manager.SetStudyStage(StudyStageRegistrationStarted)

	// TODO: save manager to repository

	// TODO: commit transaction

	// set study and manager
	s.StudyManager = &manager
	s.OnGoingStudy = study

	return nil
}

func (s *StudyService) ChangeRegistrationState(memberID, guildID, subject string, state bool) error {
	defer s.mtx.Unlock()
	s.mtx.Lock()

	manager := s.StudyManager

	// check if study is in registration stage
	if !manager.StudyStage.IsRegistrationOngoing() {
		return errors.New("발표자 등록이 불가능한 상태입니다.")
	}

	study := *s.OnGoingStudy

	// check if presentor belongs to study
	if study.GuildID != guildID {
		return errors.New("해당 디스코드 서버에서 진행중인 스터디가 아닙니다.")
	}

	// check if presentor is initialized
	member, ok := study.GetMember(memberID)
	if !ok {
		member = NewMember(memberID)
	}

	// change member's registered state
	member.SetRegistered(state)
	member.SetSubject(subject)
	study.SetMember(memberID, member)

	// TODO: save study to repository

	s.OnGoingStudy = &study

	return nil
}

func (s *StudyService) FinishRegistration(proposerID string) error {
	defer s.mtx.Unlock()
	s.mtx.Lock()

	manager := *s.StudyManager

	// check if proposer is manager
	if !manager.IsManager(proposerID) {
		return errors.New("스터디 관리자가 아닙니다.")
	}

	// check if study is in registration stage
	if !manager.StudyStage.IsRegistrationOngoing() {
		return errors.New("발표자 등록 완료가 불가능한 상태입니다.")
	}

	manager.SetStudyStage(StudyStageRegistrationFinished)

	// TODO: save manager to repository

	s.StudyManager = &manager

	return nil
}

func (s *StudyService) StartSubmission(proposerID string) error {
	defer s.mtx.Unlock()
	s.mtx.Lock()

	manager := *s.StudyManager

	// check if proposer is manager
	if !manager.IsManager(proposerID) {
		return errors.New("스터디 관리자가 아닙니다.")
	}

	// check if study can accept content submission
	if manager.StudyStage.IsRegistrationFinished() {
		return errors.New("발표 자료 제출 단계 시작이 불가능한 상태입니다.")
	}

	manager.SetStudyStage(StudyStageSubmissionStarted)

	// TODO: save manager to repository

	s.StudyManager = &manager

	return nil
}

func (s *StudyService) SubmitContent(memberID, guildID, content string) error {
	defer s.mtx.Unlock()
	s.mtx.Lock()

	manager := s.StudyManager

	// check if study can accept content submission
	if !manager.StudyStage.IsSubmissionOngoing() {
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
	member.SetContentURL(content)
	study.SetMember(memberID, member)

	// TODO: save study to repository

	s.OnGoingStudy = &study

	return nil
}

func (s *StudyService) FinishSubmission(proposerID string) error {
	defer s.mtx.Unlock()
	s.mtx.Lock()

	manager := *s.StudyManager

	// check if proposer is manager
	if !manager.IsManager(proposerID) {
		return errors.New("스터디 관리자가 아닙니다.")
	}

	// check if study can accept content submission
	if !manager.StudyStage.IsSubmissionOngoing() {
		return errors.New("발표 자료 제출 단계 종료가 불가능한 상태입니다.")
	}

	manager.SetStudyStage(StudyStageSubmissionFinished)

	// TODO: save manager to repository

	s.StudyManager = &manager

	return nil
}

func (s *StudyService) StartPresentation(proposerID string) error {
	defer s.mtx.Unlock()
	s.mtx.Lock()

	manager := *s.StudyManager

	// check if proposer is manager
	if !manager.IsManager(proposerID) {
		return errors.New("스터디 관리자가 아닙니다.")
	}

	// check if content submission is finished
	if !manager.StudyStage.IsSubmissionFinished() {
		return errors.New("발표 단계 시작이 불가능한 상태입니다.")
	}

	manager.SetStudyStage(StudyStagePresentationStarted)

	// TODO: save study manager to repository

	s.StudyManager = &manager

	return nil
}

func (s *StudyService) ChangePresentationCompletedState(proposerID, guildID, memberID string, state bool) error {
	defer s.mtx.Unlock()
	s.mtx.Lock()

	manager := *s.StudyManager

	// check if proposer is manager
	if !manager.IsManager(proposerID) {
		return errors.New("스터디 관리자가 아닙니다.")
	}

	// check if presentation is ongoing
	if !manager.StudyStage.IsPresentationOngoing() {
		return errors.New("발표 완료 상태 전환이 불가능한 상태입니다.")
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

	// set complete state
	member.SetCompleted(state)
	study.SetMember(memberID, member)

	// TODO: save study to repository

	s.OnGoingStudy = &study

	return nil
}

func (s *StudyService) FinishPresentation(proposerID string) error {
	defer s.mtx.Unlock()
	s.mtx.Lock()

	manager := *s.StudyManager

	// check if proposer is manager
	if !manager.IsManager(proposerID) {
		return errors.New("스터디 관리자가 아닙니다.")
	}

	// check if presentation is ongoing
	if !manager.StudyStage.IsPresentationOngoing() {
		return errors.New("발표 단계 종료가 불가능한 상태입니다.")
	}

	manager.SetStudyStage(StudyStagePresentationFinished)

	// TODO: save study manager to repository

	s.StudyManager = &manager

	return nil
}

func (s *StudyService) StartReview(proposerID string) error {
	defer s.mtx.Unlock()
	s.mtx.Lock()

	manager := *s.StudyManager

	// check if proposer is manager
	if !manager.IsManager(proposerID) {
		return errors.New("스터디 관리자가 아닙니다.")
	}

	// check if presentation is finished
	if !manager.StudyStage.IsPresentationFinished() {
		return errors.New("리뷰 단계 시작이 불가능한 상태입니다.")
	}

	manager.SetStudyStage(StudyStageReviewStarted)

	// TODO: save study manager to repository

	s.StudyManager = &manager

	return nil
}

func (s *StudyService) HasDoneReview(reviewerID, revieweeID string) (bool, error) {
	defer s.mtx.RUnlock()
	s.mtx.RLock()

	if !s.StudyManager.StudyStage.IsReviewOngoing() {
		return false, errors.New("리뷰 단계가 진행중이 아닙니다.")
	}

	study := s.OnGoingStudy

	// check if reviewee belongs to study
	reviewee, ok := study.GetMember(revieweeID)
	if !ok {
		return false, errors.New("활성화된 스터디에 등록되지 않은 사용자입니다.")
	}

	// check if reviewee is registered
	if !reviewee.Registered {
		return false, errors.New("발표자로 등록되지 않은 사용자입니다.")
	}

	// check if reviewee completed presentation
	if !reviewee.Completed {
		return false, errors.New("발표를 완료하지 않은 사용자입니다.")
	}

	// check if reviewer already reviewed
	return reviewee.HasDoneReview(reviewerID), nil
}

func (s *StudyService) SetReviewer(reviewerID, revieweeID string) error {
	defer s.mtx.Unlock()
	s.mtx.Lock()

	if !s.StudyManager.StudyStage.IsReviewOngoing() {
		return errors.New("리뷰 단계가 진행중이 아닙니다.")
	}

	study := *s.OnGoingStudy

	// check if reviewee belongs to study
	reviewee, ok := study.GetMember(revieweeID)
	if !ok {
		return errors.New("활성화된 스터디에 등록되지 않은 사용자입니다.")
	}

	// check if reviewee is registered
	if !reviewee.Registered {
		return errors.New("발표자로 등록되지 않은 사용자입니다.")
	}

	// check if reviewee completed presentation
	if !reviewee.Completed {
		return errors.New("발표를 완료하지 않은 사용자입니다.")
	}

	// check if reviewer already reviewed
	if reviewee.HasDoneReview(reviewerID) {
		return errors.New("이미 리뷰를 완료한 사용자입니다.")
	}

	// set reviewer
	reviewee.SetReviewer(reviewerID)
	study.SetMember(revieweeID, reviewee)

	// TODO: save study to repository

	s.OnGoingStudy = &study

	return nil
}

func (s *StudyService) FinishReview(proposerID string) error {
	defer s.mtx.Unlock()
	s.mtx.Lock()

	manager := *s.StudyManager

	// check if proposer is manager
	if !manager.IsManager(proposerID) {
		return errors.New("스터디 관리자가 아닙니다.")
	}

	// check if review is ongoing
	if !manager.StudyStage.IsReviewOngoing() {
		return errors.New("리뷰 단계 종료가 불가능한 상태입니다.")
	}

	manager.SetStudyStage(StudyStageReviewFinished)

	// TODO: save study manager to repository

	s.StudyManager = &manager

	return nil
}

func (s *StudyService) FinishStudy(proposerID string) error {
	defer s.mtx.Unlock()
	s.mtx.Lock()

	manager := *s.StudyManager

	// check if proposer is manager
	if !manager.IsManager(proposerID) {
		return errors.New("스터디 관리자가 아닙니다.")
	}

	// check if review is finished
	if !manager.StudyStage.IsReviewFinished() {
		return errors.New("스터디 종료가 불가능한 상태입니다.")
	}

	manager.SetOnGoingStudyID("")
	manager.SetStudyStage(StudyStageWait)

	// TODO: save study manager to repository

	s.StudyManager = &manager
	s.OnGoingStudy = nil

	return nil
}

// TODO: Define StudyRepository

type Repository struct{}

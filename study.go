package main

import (
	"sync"
)

type StudyState uint8

const (
	StudyStateNone     StudyState = 0
	StudyStateWait     StudyState = 1
	StudyStateRegister StudyState = 2
	StudyStateSubmit   StudyState = 3
	StudyStatePresent  StudyState = 4
	StudyStateReview   StudyState = 5
)

func (s StudyState) String() string {
	switch s {
	case StudyStateWait:
		return "다음 주차 대기"
	case StudyStateRegister:
		return "발표자 등록"
	case StudyStateSubmit:
		return "발표자료 제출"
	case StudyStatePresent:
		return "발표"
	case StudyStateReview:
		return "리뷰 및 피드백"
	default:
		return "몰?루"
	}
}

type StudyManager struct {
	GuildID         string
	NoticeChannelID string

	ManagerID     string
	SubManagerIDs []string

	OnGoingStudyID string
	StudyState     StudyState

	mtx *sync.Mutex
}

func NewStudyManager(guildID string, ManagerID string) *StudyManager {
	return &StudyManager{
		GuildID:         guildID,
		NoticeChannelID: "",
		ManagerID:       ManagerID,
		SubManagerIDs:   []string{},
		OnGoingStudyID:  "",
		StudyState:      StudyStateNone,
		mtx:             &sync.Mutex{},
	}
}

func (s *StudyManager) SetNoticeChannelID(channelID string) {
	defer s.mtx.Unlock()
	s.mtx.Lock()
	s.NoticeChannelID = channelID
}

func (s *StudyManager) IsManager(userID string) bool {
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
		}
	}
}

func (s *StudyManager) IsSubManager(userID string) bool {
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

func (s *StudyManager) SetStudyState(state StudyState) {
	defer s.mtx.Unlock()
	s.mtx.Lock()
	s.StudyState = state
}

type Member struct {
	Nickname   string
	Registered bool

	Subject    string
	ContentURL string
	Completed  bool

	mtx *sync.Mutex
}

func NewMember(nickname string) Member {
	return Member{
		Nickname:   nickname,
		Registered: false,
		Subject:    "",
		ContentURL: "",
		Completed:  false,
		mtx:        &sync.Mutex{},
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

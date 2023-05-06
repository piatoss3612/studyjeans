package main

import (
	"sync"
)

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

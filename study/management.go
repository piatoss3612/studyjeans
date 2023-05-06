package study

import "sync"

type Management struct {
	GuildID         string `bson:"_id"`
	NoticeChannelID string `bson:"notice_channel_id"`

	ManagerID string `bson:"manager_id"`

	OnGoingStudyID    string     `bson:"on_going_study_id"`
	CurrentStudyStage StudyStage `bson:"current_study_stage"`

	mtx *sync.RWMutex
}

func NewManagement() *Management {
	return &Management{
		GuildID:           "",
		NoticeChannelID:   "",
		ManagerID:         "",
		OnGoingStudyID:    "",
		CurrentStudyStage: StudyStageNone,
		mtx:               &sync.RWMutex{},
	}
}

func (s *Management) SetGuildID(guildID string) {
	defer s.mtx.Unlock()
	s.mtx.Lock()

	s.GuildID = guildID
}

func (s *Management) SetNoticeChannelID(channelID string) {
	defer s.mtx.Unlock()
	s.mtx.Lock()

	s.NoticeChannelID = channelID
}

func (s *Management) SetManagerID(userID string) {
	defer s.mtx.Unlock()
	s.mtx.Lock()

	s.ManagerID = userID
}

func (s *Management) IsManager(userID string) bool {
	defer s.mtx.RLock()
	s.mtx.RLock()

	return s.ManagerID == userID
}

func (s *Management) SetOnGoingStudyID(studyID string) {
	defer s.mtx.Unlock()
	s.mtx.Lock()

	s.OnGoingStudyID = studyID
}

func (s *Management) SetCurrentStudyStage(state StudyStage) {
	defer s.mtx.Unlock()
	s.mtx.Lock()

	s.CurrentStudyStage = state
}

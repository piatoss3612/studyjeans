package study

import (
	"sync"
	"time"
)

type Management struct {
	ID              string `bson:"_id,omitempty"`
	GuildID         string `bson:"guild_id"`
	NoticeChannelID string `bson:"notice_channel_id"`

	ManagerID string `bson:"manager_id"`

	OnGoingStudyID    string     `bson:"ongoing_study_id"`
	CurrentStudyStage StudyStage `bson:"current_study_stage"`

	CreatedAt time.Time `bson:"created_at"`
	UpdatedAt time.Time `bson:"updated_at"`

	mtx *sync.RWMutex
}

func NewManagement() *Management {
	return &Management{
		GuildID:           "",
		NoticeChannelID:   "",
		ManagerID:         "",
		OnGoingStudyID:    "",
		CurrentStudyStage: StudyStageNone,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
		mtx:               &sync.RWMutex{},
	}
}

func (s *Management) SetID(id string) {
	defer s.mtx.Unlock()
	s.mtx.Lock()

	s.ID = id
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

func (s *Management) SetUpdatedAt(t time.Time) {
	defer s.mtx.Unlock()
	s.mtx.Lock()

	s.UpdatedAt = t
}

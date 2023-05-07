package study

import (
	"time"
)

type Management struct {
	ID              string `bson:"_id,omitempty"`
	GuildID         string `bson:"guild_id"`
	NoticeChannelID string `bson:"notice_channel_id"`

	ManagerID string `bson:"manager_id"`

	OngoingStudyID    string     `bson:"ongoing_study_id"`
	CurrentStudyStage StudyStage `bson:"current_study_stage"`

	CreatedAt time.Time `bson:"created_at"`
	UpdatedAt time.Time `bson:"updated_at"`
}

func NewManagement() *Management {
	return &Management{
		GuildID:           "",
		NoticeChannelID:   "",
		ManagerID:         "",
		OngoingStudyID:    "",
		CurrentStudyStage: StudyStageNone,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}
}

func (s *Management) SetID(id string) {
	s.ID = id
}

func (s *Management) SetGuildID(guildID string) {
	s.GuildID = guildID
}

func (s *Management) SetNoticeChannelID(channelID string) {
	s.NoticeChannelID = channelID
}

func (s *Management) SetManagerID(userID string) {
	s.ManagerID = userID
}

func (s *Management) IsManager(userID string) bool {
	return s.ManagerID == userID
}

func (s *Management) SetOngoingStudyID(studyID string) {
	s.OngoingStudyID = studyID
}

func (s *Management) SetCurrentStudyStage(state StudyStage) {
	s.CurrentStudyStage = state
}

func (s *Management) SetUpdatedAt(t time.Time) {
	s.UpdatedAt = t
}

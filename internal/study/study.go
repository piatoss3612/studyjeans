package study

import (
	"time"
)

type Study struct {
	ID                  string `bson:"_id,omitempty"`
	GuildID             string `bson:"guild_id"`
	NoticeChannelID     string `bson:"notice_channel_id"`
	ReflectionChannelID string `bson:"reflection_channel_id"`
	ManagerID           string `bson:"manager_id"`
	OngoingRoundID      string `bson:"ongoing_round_id"`
	SpreadsheetURL      string `bson:"spreadsheet_url"`
	CurrentStage        Stage  `bson:"current_stage"`
	TotalRound          int8   `bson:"total_round"`

	CreatedAt time.Time `bson:"created_at"`
	UpdatedAt time.Time `bson:"updated_at"`
}

func New() Study {
	return Study{
		GuildID:             "",
		NoticeChannelID:     "",
		ReflectionChannelID: "",
		ManagerID:           "",
		OngoingRoundID:      "",
		SpreadsheetURL:      "",
		CurrentStage:        StageNone,
		CreatedAt:           time.Now(),
		UpdatedAt:           time.Now(),
	}
}

func (s *Study) SetID(id string) {
	s.ID = id
}

func (s *Study) SetGuildID(guildID string) {
	s.GuildID = guildID
}

func (s *Study) SetNoticeChannelID(channelID string) {
	s.NoticeChannelID = channelID
}

func (s *Study) SetReflectionChannelID(channelID string) {
	s.ReflectionChannelID = channelID
}

func (s *Study) SetManagerID(userID string) {
	s.ManagerID = userID
}

func (s *Study) IsManager(userID string) bool {
	return s.ManagerID == userID
}

func (s *Study) SetOngoingRoundID(roundID string) {
	s.OngoingRoundID = roundID
}

func (s *Study) SetSpreadsheetURL(url string) {
	s.SpreadsheetURL = url
}

func (s *Study) SetCurrentStage(state Stage) {
	s.CurrentStage = state
}

func (s *Study) IncrementTotalRound() {
	s.TotalRound++
}

func (s *Study) SetUpdatedAt(t time.Time) {
	s.UpdatedAt = t
}

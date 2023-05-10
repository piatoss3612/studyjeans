package study

import (
	"time"
)

type Round struct {
	ID      string `bson:"_id,omitempty"`
	GuildID string `bson:"guild_id"`

	Number     int8              `bson:"number"`
	Title      string            `bson:"title"`
	ContentURL string            `bson:"content_url"`
	Stage      Stage             `bson:"stage"`
	Members    map[string]Member `bson:"members"`

	CreatedAt time.Time `bson:"created_at"`
	UpdatedAt time.Time `bson:"updated_at"`
}

func NewRound() Round {
	return Round{
		ID:        "",
		GuildID:   "",
		Title:     "",
		Members:   map[string]Member{},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

func (s *Round) SetID(id string) {
	s.ID = id
}

func (s *Round) SetGuildID(guildID string) {
	s.GuildID = guildID
}

func (s *Round) SetNumber(number int8) {
	s.Number = number
}

func (s *Round) SetTitle(title string) {
	s.Title = title
}

func (s *Round) SetContentURL(contentURL string) {
	s.ContentURL = contentURL
}

func (s *Round) SetMember(memberID string, member Member) {
	s.Members[memberID] = member
}

func (s *Round) GetMember(memberID string) (Member, bool) {
	member, ok := s.Members[memberID]
	return member, ok
}

func (s *Round) GetMembers() []Member {
	members := []Member{}
	for _, v := range s.Members {
		members = append(members, v)
	}
	return members
}

func (s *Round) SetUpdatedAt(updatedAt time.Time) {
	s.UpdatedAt = updatedAt
}

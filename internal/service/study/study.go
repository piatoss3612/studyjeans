package study

import (
	"time"
)

type Study struct {
	ID      string `bson:"_id,omitempty"`
	GuildID string `bson:"guild_id"`

	Title     string            `bson:"title"`
	Members   map[string]Member `bson:"members"`
	CreatedAt time.Time         `bson:"created_at"`
	UpdatedAt time.Time         `bson:"updated_at"`
}

func New() Study {
	return Study{
		ID:        "",
		GuildID:   "",
		Title:     "",
		Members:   map[string]Member{},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

func (s *Study) SetID(id string) {
	s.ID = id
}

func (s *Study) SetGuildID(guildID string) {
	s.GuildID = guildID
}

func (s *Study) SetTitle(title string) {
	s.Title = title
}

func (s *Study) SetMember(memberID string, member Member) {
	s.Members[memberID] = member
}

func (s *Study) GetMember(memberID string) (Member, bool) {
	member, ok := s.Members[memberID]
	return member, ok
}

func (s *Study) GetMembers() []Member {
	members := []Member{}
	for _, v := range s.Members {
		members = append(members, v)
	}
	return members
}

func (s *Study) SetUpdatedAt(updatedAt time.Time) {
	s.UpdatedAt = updatedAt
}

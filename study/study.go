package study

import (
	"sync"
	"time"
)

type Study struct {
	ID      string `bson:"_id"`
	GuildID string `bson:"guild_id"`

	Title     string            `bson:"title"`
	Members   map[string]Member `bson:"members"`
	CreatedAt time.Time         `bson:"created_at"`
	UpdatedAt time.Time         `bson:"updated_at"`

	mtx *sync.RWMutex
}

func New() *Study {
	return &Study{
		ID:        "",
		GuildID:   "",
		Title:     "",
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

func (s *Study) SetGuildID(guildID string) {
	defer s.mtx.Unlock()
	s.mtx.Lock()

	s.GuildID = guildID
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

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

func (r *Round) SetID(id string) {
	r.ID = id
}

func (r *Round) SetGuildID(guildID string) {
	r.GuildID = guildID
}

func (r *Round) SetNumber(number int8) {
	r.Number = number
}

func (r *Round) SetTitle(title string) {
	r.Title = title
}

func (r *Round) SetContentURL(contentURL string) {
	r.ContentURL = contentURL
}

func (r *Round) SetStage(stage Stage) {
	r.Stage = stage
}

func (r *Round) SetMember(memberID string, member Member) {
	r.Members[memberID] = member
}

func (r *Round) GetMember(memberID string) (Member, bool) {
	member, ok := r.Members[memberID]
	return member, ok
}

func (r *Round) GetMembers() []Member {
	members := []Member{}
	for _, v := range r.Members {
		members = append(members, v)
	}
	return members
}

func (r *Round) SetUpdatedAt(updatedAt time.Time) {
	r.UpdatedAt = updatedAt
}

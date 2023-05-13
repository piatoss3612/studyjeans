package study

import (
	"time"
)

type Round struct {
	ID      string `bson:"_id,omitempty" json:"id,omitempty"`
	GuildID string `bson:"guild_id" json:"guild_id,omitempty"`

	Number     int8              `bson:"number" json:"number"`
	Title      string            `bson:"title" json:"title"`
	ContentURL string            `bson:"content_url" json:"content_url"`
	Stage      Stage             `bson:"stage" json:"stage"`
	Members    map[string]Member `bson:"members" json:"members"`

	CreatedAt time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time `bson:"updated_at" json:"updated_at"`
}

func NewRound() Round {
	return Round{
		ID:        "",
		GuildID:   "",
		Number:    0,
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

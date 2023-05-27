package logger

import (
	"time"
)

type Round struct {
	Number     int8              `json:"number"`
	Title      string            `json:"title"`
	ContentURL string            `json:"content_url"`
	Stage      Stage             `json:"stage"`
	Members    map[string]Member `json:"members"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func NewRound() Round {
	return Round{
		Number:    0,
		Title:     "",
		Members:   map[string]Member{},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

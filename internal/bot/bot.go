package bot

import (
	"github.com/bwmarrin/discordgo"
	"github.com/piatoss3612/studyjeans/pkg/command"
)

type Bot struct {
	m command.CommandManager
	s *discordgo.Session
	r *command.CommandRegistry
}

func New(s *discordgo.Session) *Bot {
	return &Bot{s: s}
}

func (b *Bot) AddHandler(h interface{}) {
	b.s.AddHandler(h)
}

func (b *Bot) Open() error {
	return b.s.Open()
}

func (b *Bot) Close() error {
	return b.s.Close()
}

package bot

import "github.com/bwmarrin/discordgo"

type Bot struct {
	s *discordgo.Session
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

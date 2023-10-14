package bot

import "github.com/bwmarrin/discordgo"

type Bot interface {
	AddHandler(h interface{})
	Open() error
	Close() error
}

type DiscordBot struct {
	s *discordgo.Session
}

func NewDiscordBot(token string) (*DiscordBot, error) {
	s, err := discordgo.New("Bot " + token)
	if err != nil {
		return nil, err
	}

	return &DiscordBot{s: s}, nil
}

func (b *DiscordBot) AddHandler(h interface{}) {
	b.s.AddHandler(h)
}

func (b *DiscordBot) Open() error {
	return b.s.Open()
}

func (b *DiscordBot) Close() error {
	return b.s.Close()
}

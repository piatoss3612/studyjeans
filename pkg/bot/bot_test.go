package bot

import (
	"os"
	"testing"

	"github.com/bwmarrin/discordgo"
	_ "github.com/joho/godotenv/autoload"
)

func TestNewDiscordBot(t *testing.T) {
	s := &discordgo.Session{}
	bot := New(s)
	if bot.s != s {
		t.Errorf("Expected %v, got %v", s, bot.s)
	}
}

func TestDiscordBot_AddHandler(t *testing.T) {
	s := &discordgo.Session{}
	bot := New(s)
	bot.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {})
}

func TestDiscordBot_Open(t *testing.T) {
	s, err := discordgo.New("Bot " + os.Getenv("BOT_TOKEN"))
	if err != nil {
		t.Errorf("Expected nil, got %v", err)
	}

	b := New(s)

	if err := b.Open(); err != nil {
		t.Errorf("Expected nil, got %v", err)
	}
}

func TestDiscordBot_Close(t *testing.T) {
	s, err := discordgo.New("Bot " + os.Getenv("BOT_TOKEN"))
	if err != nil {
		t.Errorf("Expected nil, got %v", err)
	}

	b := New(s)

	if err := b.Close(); err != nil {
		t.Errorf("Expected nil, got %v", err)
	}
}

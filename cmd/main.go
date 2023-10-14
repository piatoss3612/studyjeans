package main

import (
	"os"

	"github.com/bwmarrin/discordgo"
	_ "github.com/joho/godotenv/autoload"
	bh "github.com/piatoss3612/studyjeans/internal/bot"
	"github.com/piatoss3612/studyjeans/internal/commands/ping"
	"github.com/piatoss3612/studyjeans/pkg/bot"
	"github.com/piatoss3612/studyjeans/pkg/command"
	"github.com/piatoss3612/studyjeans/pkg/shutdown"
	"go.uber.org/zap"
)

func main() {
	l, _ := zap.NewProduction(zap.AddCaller(), zap.Fields(zap.String("service", "studyjeans")))
	defer l.Sync()

	s, err := discordgo.New("Bot " + os.Getenv("BOT_TOKEN"))
	if err != nil {
		panic(err)
	}

	s.Identify.Intents = discordgo.IntentGuildMembers | discordgo.IntentGuildMessages |
		discordgo.IntentGuilds | discordgo.IntentDirectMessages

	r := command.NewCommandRegistry()

	r.RegisterCommand(ping.New())

	m := command.NewCommandManager(s, r)

	b := bot.New(s)
	b.AddHandler(bh.NewApplicationCommandHandler(m, l))

	if err := b.Open(); err != nil {
		panic(err)
	}

	l.Info("Studyjeans bot is now running. Press CTRL-C to exit.")

	<-shutdown.GracefulShutdown(func() error {
		return b.Close()
	}, os.Interrupt)

	l.Info("Successfully shutdown studyjeans bot!")
}

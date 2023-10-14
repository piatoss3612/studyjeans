package main

import (
	"fmt"
	"os"

	"github.com/bwmarrin/discordgo"
	_ "github.com/joho/godotenv/autoload"
	"github.com/piatoss3612/studyjeans/pkg/bot"
	"github.com/piatoss3612/studyjeans/pkg/shutdown"
)

func main() {
	s, err := discordgo.New("Bot " + os.Getenv("BOT_TOKEN"))
	if err != nil {
		panic(err)
	}

	b := bot.New(s)

	if err := b.Open(); err != nil {
		panic(err)
	}

	<-shutdown.GracefulShutdown(func() error {
		return b.Close()
	})

	fmt.Println("Gracefully shutdown")
}

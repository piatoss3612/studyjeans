package main

import (
	"fmt"
	"os"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
	"github.com/piatoss3612/studyjeans/pkg/bot"
	"github.com/piatoss3612/studyjeans/pkg/shutdown"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		panic(err)
	}

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

package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
	_ "github.com/joho/godotenv/autoload"
)

func main() {
	defer func() {
		if r := recover(); r != nil {
			log.Println("Recovered:", r)
		}
	}()

	session, err := discordgo.New("Bot " + os.Getenv("BOT_TOKEN"))
	if err != nil {
		log.Fatal(err)
	}

	if err = session.Open(); err != nil {
		log.Fatal(err)
	}

	log.Println("Bot is running!")

	stop := make(chan struct{})
	shutdown := make(chan os.Signal, 1)

	signal.Notify(shutdown,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT,
	)

	go func() {
		defer func() {
			close(shutdown)
			close(stop)
		}()
		<-shutdown
	}()

	<-stop

	if err = session.Close(); err != nil {
		log.Fatal(err)
	}

	log.Println("Bot is stopped!")
}

package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"context"

	"github.com/bwmarrin/discordgo"
	_ "github.com/joho/godotenv/autoload"
	"github.com/piatoss3612/presentation-helper-bot/study"
)

func main() {
	defer func() {
		if r := recover(); r != nil {
			log.Println("Recovered:", r)
		}
	}()

	cfg, err := NewConfig(os.Getenv("CONFIG_FILE"))
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	mongoClient, err := ConnectMongoDB(ctx, cfg.MongoURI)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err = mongoClient.Disconnect(context.Background()); err != nil {
			log.Fatal(err)
		}
		log.Println("Disconnected from MongoDB!")
	}()

	log.Println("Connected to MongoDB!")

	tx := study.NewTx(mongoClient)

	svc, err := study.NewService(tx, cfg.GuildID)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Study service is ready!")

	session, err := discordgo.New("Bot " + cfg.BotToken)
	if err != nil {
		log.Fatal(err)
	}

	session.Identify.Intents = discordgo.IntentGuildMembers | discordgo.IntentGuildMessages | discordgo.IntentGuilds | discordgo.IntentDirectMessages

	session.AddHandler(ready)

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

func ready(s *discordgo.Session, event *discordgo.Ready) {
	s.UpdateGameStatus(0, "발표 준비")
}

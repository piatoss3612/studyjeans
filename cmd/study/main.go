package main

import (
	"log"
	"os"
	"time"

	"context"

	"github.com/bwmarrin/discordgo"
	_ "github.com/joho/godotenv/autoload"
	"github.com/piatoss3612/presentation-helper-bot/internal/db"
	"github.com/piatoss3612/presentation-helper-bot/internal/service/study"
)

func main() {
	defer func() {
		if r := recover(); r != nil {
			log.Println("Recovered:", r)
		}
	}()

	mustSetTimezone(os.Getenv("TIME_ZONE"))

	run()
}

func run() {
	cfg, err := NewConfig(os.Getenv("CONFIG_FILE"))
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	mongoClient, err := db.ConnectMongoDB(ctx, cfg.MongoURI)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		_ = mongoClient.Disconnect(context.Background())
		log.Println("Disconnected from MongoDB!")
	}()

	log.Println("Connected to MongoDB!")

	svc, err := study.NewService(ctx, study.NewTx(mongoClient, cfg.DBName),
		cfg.GuildID, cfg.ManagerID, cfg.NoticeChannelID)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Study service is ready!")

	sess, err := discordgo.New("Bot " + cfg.BotToken)
	if err != nil {
		log.Fatal(err)
	}

	bot := NewStudyBot(sess, svc).Setup()

	stop, err := bot.Run()
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		_ = bot.Close()
		log.Println("Disconnected from Discord!")
	}()

	log.Println("Connected to Discord!")

	<-stop
}

func mustSetTimezone(tz string) {
	loc, err := time.LoadLocation(tz)
	if err != nil {
		log.Fatal(err)
	}

	time.Local = loc
}

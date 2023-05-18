package main

import (
	"log"
	"os"
	"time"

	"context"

	"github.com/bwmarrin/discordgo"
	_ "github.com/joho/godotenv/autoload"
	app "github.com/piatoss3612/presentation-helper-bot/internal/app/study"
	"github.com/piatoss3612/presentation-helper-bot/internal/config"
	"github.com/piatoss3612/presentation-helper-bot/internal/msgqueue"
	service "github.com/piatoss3612/presentation-helper-bot/internal/service/study"
	store "github.com/piatoss3612/presentation-helper-bot/internal/store/study"
	"github.com/piatoss3612/presentation-helper-bot/internal/tools"
	"go.uber.org/zap"
)

var sugar *zap.SugaredLogger

func main() {
	logger, _ := zap.NewProduction()
	defer func() {
		_ = logger.Sync()
	}()

	sugar = logger.Sugar()

	defer func() {
		if r := recover(); r != nil {
			sugar.Info("Panic recovered", "error", r)
		}
	}()

	mustSetTimezone(os.Getenv("TIME_ZONE"))

	run()
}

func run() {
	cfg := mustLoadConfig(os.Getenv("CONFIG_FILE"))

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	tx, close := mustInitTx(ctx, cfg.MongoDB.URI, cfg.MongoDB.DBName)
	defer func() {
		_ = close()
		sugar.Info("Disconnected from MongoDB!")
	}()

	sugar.Info("Connected to MongoDB!")

	cache := mustInitStudyCache(ctx, cfg.Redis.Addr, 1*time.Minute)

	sugar.Info("Study cache is ready!")

	pub, close := mustInitPublisher(ctx, cfg.RabbitMQ.Addr, cfg.RabbitMQ.Exchange, cfg.RabbitMQ.Kind)
	defer func() {
		_ = close()
		sugar.Info("Disconnected from RabbitMQ!")
	}()

	sugar.Info("Study/Event publisher is ready!")

	svc := mustInitStudyService(ctx, tx, cfg.Discord.GuildID,
		cfg.Discord.ManagerID, cfg.Discord.NoticeChannelID, cfg.Discord.ReflectionChannelID)

	sugar.Info("Study service is ready!")

	sess := mustOpenDiscordSession(cfg.Discord.BotToken)

	bot := app.New(sess, svc, cache, pub, sugar).Setup()

	stop, err := bot.Run()
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		_ = bot.Close()
		sugar.Info("Disconnected from Discord!")
	}()

	sugar.Info("Connected to Discord!")

	<-stop
}

func mustLoadConfig(path string) *config.StudyConfig {
	cfg, err := config.NewStudyConfig(path)
	if err != nil {
		sugar.Fatal(err)
	}

	return cfg
}

func mustInitTx(ctx context.Context, uri, dbname string) (store.Tx, func() error) {
	mongoClient, err := tools.ConnectMongoDB(ctx, uri)
	if err != nil {
		sugar.Fatal(err)
	}

	return store.NewMongoTx(mongoClient, store.WithDBName(dbname)), func() error { return mongoClient.Disconnect(context.Background()) }
}

func mustInitStudyCache(ctx context.Context, addr string, ttl time.Duration) store.Cache {
	cache, err := tools.ConnectRedisCache(ctx, addr, ttl)
	if err != nil {
		sugar.Fatal(err)
	}

	return store.NewCache(cache)
}

func mustInitPublisher(ctx context.Context, addr, exchange, kind string) (msgqueue.Publisher, func() error) {
	rabbit := <-tools.RedialRabbitMQ(ctx, addr)

	if rabbit == nil {
		sugar.Fatal("Failed to connect to RabbitMQ")
	}

	pub, err := msgqueue.NewPublisher(rabbit, exchange, kind)
	if err != nil {
		sugar.Fatal(err)
	}

	return pub, func() error { return rabbit.Close() }
}

func mustInitStudyService(ctx context.Context, tx store.Tx, guildID, managerID, noticeChID, reflectionChID string) service.Service {
	svc, err := service.New(ctx, tx, guildID, managerID, noticeChID, reflectionChID)
	if err != nil {
		sugar.Fatal(err)
	}

	return svc
}

func mustOpenDiscordSession(token string) *discordgo.Session {
	sess, err := discordgo.New("Bot " + token)
	if err != nil {
		sugar.Fatal(err)
	}

	return sess
}

func mustSetTimezone(tz string) {
	loc, err := time.LoadLocation(tz)
	if err != nil {
		sugar.Fatal(err)
	}

	time.Local = loc
}

package main

import (
	"log"
	"os"
	"time"

	"context"

	"github.com/bwmarrin/discordgo"
	_ "github.com/joho/godotenv/autoload"
	"github.com/piatoss3612/presentation-helper-bot/internal/bot"
	"github.com/piatoss3612/presentation-helper-bot/internal/bot/command"
	"github.com/piatoss3612/presentation-helper-bot/internal/bot/command/profile"
	"github.com/piatoss3612/presentation-helper-bot/internal/bot/command/submit"
	"github.com/piatoss3612/presentation-helper-bot/internal/bot/help"
	"github.com/piatoss3612/presentation-helper-bot/internal/config"
	"github.com/piatoss3612/presentation-helper-bot/internal/event/msgqueue"
	"github.com/piatoss3612/presentation-helper-bot/internal/study/cache"
	"github.com/piatoss3612/presentation-helper-bot/internal/study/cache/redis"
	"github.com/piatoss3612/presentation-helper-bot/internal/study/repository"
	"github.com/piatoss3612/presentation-helper-bot/internal/study/repository/mongo"
	"github.com/piatoss3612/presentation-helper-bot/internal/study/service"
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

	svc := service.New(tx)
	sugar.Info("Study service is ready!")

	sess := mustOpenDiscordSession(cfg.Discord.BotToken)

	cmdReg := command.NewRegisterer()

	helpCmd := help.NewHelpCommand(pub, sugar)
	helpCmd.Register(cmdReg)

	profileCmd := profile.NewProfileCommand(pub, sugar)
	profileCmd.Register(cmdReg)

	submitCmd := submit.NewSubmitCommand(svc, pub, sugar)
	submitCmd.Register(cmdReg)

	b := bot.New(cmdReg, sess, sugar).Setup()

	stop, err := b.Run()
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		_ = b.Close()
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

func mustInitTx(ctx context.Context, uri, dbname string) (repository.Tx, func() error) {
	mongoClient, err := tools.ConnectMongoDB(ctx, uri)
	if err != nil {
		sugar.Fatal(err)
	}

	return mongo.NewMongoTx(mongoClient, mongo.WithDBName(dbname)), func() error { return mongoClient.Disconnect(context.Background()) }
}

func mustInitStudyCache(ctx context.Context, addr string, ttl time.Duration) cache.Cache {
	cache, err := tools.ConnectRedisCache(ctx, addr, ttl)
	if err != nil {
		sugar.Fatal(err)
	}

	return redis.NewCache(cache)
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

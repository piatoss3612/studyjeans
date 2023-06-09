package main

import (
	"log"
	"os"
	"time"

	"context"

	"github.com/bwmarrin/discordgo"
	_ "github.com/joho/godotenv/autoload"
	"github.com/piatoss3612/my-study-bot/internal/bot"
	"github.com/piatoss3612/my-study-bot/internal/bot/command"
	"github.com/piatoss3612/my-study-bot/internal/bot/command/admin"
	"github.com/piatoss3612/my-study-bot/internal/bot/command/feedback"
	"github.com/piatoss3612/my-study-bot/internal/bot/command/help"
	"github.com/piatoss3612/my-study-bot/internal/bot/command/info"
	"github.com/piatoss3612/my-study-bot/internal/bot/command/profile"
	"github.com/piatoss3612/my-study-bot/internal/bot/command/reflection"
	"github.com/piatoss3612/my-study-bot/internal/bot/command/registration"
	"github.com/piatoss3612/my-study-bot/internal/bot/command/submit"
	"github.com/piatoss3612/my-study-bot/internal/cache"
	"github.com/piatoss3612/my-study-bot/internal/cache/redis"
	"github.com/piatoss3612/my-study-bot/internal/config"
	"github.com/piatoss3612/my-study-bot/internal/pubsub"
	"github.com/piatoss3612/my-study-bot/internal/pubsub/rabbitmq"
	"github.com/piatoss3612/my-study-bot/internal/study/repository"
	"github.com/piatoss3612/my-study-bot/internal/study/repository/mongo"
	"github.com/piatoss3612/my-study-bot/internal/study/service"
	"github.com/piatoss3612/my-study-bot/internal/utils"
	"go.uber.org/zap"
)

var sugar *zap.SugaredLogger

func main() {
	logger, _ := zap.NewProduction(zap.Fields(zap.String("service", "study-bot")))
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

	tx, txClose := mustInitTx(ctx, cfg.MongoDB.URI, cfg.MongoDB.DBName)
	defer func() {
		_ = txClose()
		sugar.Info("Disconnected from MongoDB!")
	}()

	sugar.Info("Connected to MongoDB!")

	cache := mustInitStudyCache(ctx, cfg.Redis.Addr, 1*time.Minute)

	sugar.Info("Study cache is ready!")

	pub, pubClose := mustInitPublisher(ctx, cfg.RabbitMQ.Addr, cfg.RabbitMQ.Exchange, cfg.RabbitMQ.Kind)
	defer func() {
		_ = pubClose()
		sugar.Info("Disconnected from RabbitMQ!")
	}()

	sugar.Info("Study/Event publisher is ready!")

	svc := service.New(tx)
	sugar.Info("Study service is ready!")

	cmdReg := registerCommands(svc, pub, cache)
	handler := command.NewHandler(cmdReg.HandleFuncs())

	b := bot.New(mustOpenDiscordSession(cfg.Discord.BotToken), sugar)

	stop, err := b.Run()
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		_ = b.Close()
		sugar.Info("Disconnected from Discord!")
	}()

	sugar.Info("Connected to Discord!")

	b.RegisterHandler(handler)

	if err := b.RegisterCommands(cmdReg.Commands()); err != nil {
		sugar.Fatal(err)
	}
	defer func() {
		_ = b.RemoveCommands()
		sugar.Info("Removed commands!")
	}()

	sugar.Info("Registered commands!")

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
	mongoClient, err := utils.ConnectMongoDB(ctx, uri)
	if err != nil {
		sugar.Fatal(err)
	}

	return mongo.NewMongoTx(mongoClient, mongo.WithDBName(dbname)), func() error { return mongoClient.Disconnect(context.Background()) }
}

func mustInitStudyCache(ctx context.Context, addr string, ttl time.Duration) cache.Cache {
	cache, err := utils.ConnectRedisCache(ctx, addr, ttl)
	if err != nil {
		sugar.Fatal(err)
	}

	return redis.NewCache(cache)
}

func mustInitPublisher(ctx context.Context, addr, exchange, kind string) (pubsub.Publisher, func() error) {
	rabbit := <-utils.RedialRabbitMQ(ctx, addr)

	if rabbit == nil {
		sugar.Fatal("Failed to connect to RabbitMQ")
	}

	pub, err := rabbitmq.NewPublisher(rabbit, exchange, kind)
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

func registerCommands(svc service.Service, pub pubsub.Publisher, cache cache.Cache) command.Registerer {
	reg := command.NewRegisterer()

	admin.NewAdminCommand(svc, pub, sugar).Register(reg)
	help.NewHelpCommand().Register(reg)
	profile.NewProfileCommand(sugar).Register(reg)
	info.NewInfoCommand(svc, cache).Register(reg)
	registration.NewRegistrationCommand(svc).Register(reg)
	submit.NewSubmitCommand(svc).Register(reg)
	feedback.NewFeedbackCommand(svc).Register(reg)
	reflection.NewReflectionCommand(svc).Register(reg)

	return reg
}

func mustSetTimezone(tz string) {
	loc, err := time.LoadLocation(tz)
	if err != nil {
		sugar.Fatal(err)
	}

	time.Local = loc
}

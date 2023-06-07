package main

import (
	"context"
	"log"
	"os"
	"strconv"
	"time"

	_ "github.com/joho/godotenv/autoload"
	"github.com/piatoss3612/my-study-bot/internal/config"
	cerrors "github.com/piatoss3612/my-study-bot/internal/errors"
	eevent "github.com/piatoss3612/my-study-bot/internal/errors/event"
	"github.com/piatoss3612/my-study-bot/internal/logger/service"
	"github.com/piatoss3612/my-study-bot/internal/pubsub"
	"github.com/piatoss3612/my-study-bot/internal/pubsub/rabbitmq"
	"github.com/piatoss3612/my-study-bot/internal/study"
	sevent "github.com/piatoss3612/my-study-bot/internal/study/event"
	"github.com/piatoss3612/my-study-bot/internal/utils"
	"go.uber.org/zap"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

var sugar *zap.SugaredLogger

func main() {
	l, _ := zap.NewProduction(zap.Fields(zap.String("service", "study-logger")))
	defer func() {
		_ = l.Sync()
	}()

	sugar = l.Sugar()

	defer func() {
		if r := recover(); r != nil {
			sugar.Info("Panic recovered", "error", r)
		}
	}()

	mustSetTimezone(os.Getenv("TIME_ZONE"))

	run()
}

func run() {
	cfg := mustLoadConfig(os.Getenv("LOGGER_CONFIG_FILE"))

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	sub, close := mustInitSubscriber(ctx, cfg.RabbitMQ.Addr, cfg.RabbitMQ.Exchange, cfg.RabbitMQ.Kind, cfg.RabbitMQ.Queue)
	defer func() {
		_ = close()
		sugar.Info("RabbitMQ connection is closed!")
	}()

	sheetsSvc := mustInitSheetsService(ctx)

	sugar.Info("Sheets service is ready!")

	sh := mustInitStudyEventHandler(ctx, sheetsSvc)
	eh := mustInitErrorEventHandler(ctx, sheetsSvc)

	mapper := pubsub.NewMapper()

	type mapping struct {
		topic string
		h     pubsub.Handler
	}

	mappings := []mapping{
		{topic: study.EventTopicStudyRoundCreated.String(), h: sh},
		{topic: study.EventTopicStudyRoundFinished.String(), h: sh},
		{topic: study.EventTopicStudyRoundProgress.String(), h: sh},
		{topic: cerrors.EventTopicError.String(), h: eh},
	}

	topics := make([]string, 0, len(mappings))

	for _, m := range mappings {
		mapper.Register(m.topic, m.h)
		topics = append(topics, m.topic)
	}

	sugar.Info("Event handlers are ready!")

	svc := service.New(sub, mapper, sugar)

	stop := svc.Run()

	sugar.Info("Logger service is running!")

	svc.Listen(stop, topics)
}

func mustLoadConfig(path string) *config.LoggerConfig {
	cfg, err := config.NewLoggerConfig(path)
	if err != nil {
		sugar.Fatal(err)
	}

	return cfg
}

func mustInitSubscriber(ctx context.Context, addr, exchange, kind, queue string) (pubsub.Subscriber, func() error) {
	rabbit := <-utils.RedialRabbitMQ(ctx, addr)

	if rabbit == nil {
		sugar.Fatal("Failed to connect to RabbitMQ")
	}

	sub, err := rabbitmq.NewSubscriber(rabbit, exchange, kind, queue)
	if err != nil {
		log.Println(err)
		sugar.Fatal(err)
	}

	return sub, func() error { return rabbit.Close() }
}

func mustInitSheetsService(ctx context.Context) *sheets.Service {
	b, err := os.ReadFile(os.Getenv("SHEETS_CREDENTIALS"))
	if err != nil {
		sugar.Fatalf("Unable to read client secret file: %v", err)
	}

	config, err := google.JWTConfigFromJSON(b, "https://www.googleapis.com/auth/spreadsheets")
	if err != nil {
		sugar.Fatalf("Unable to parse client secret file to config: %v", err)
	}

	client := config.Client(ctx)

	srv, err := sheets.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		sugar.Fatalf("Unable to retrieve Sheets client: %v", err)
	}

	return srv
}

func mustInitStudyEventHandler(ctx context.Context, s *sheets.Service) pubsub.Handler {
	progressSheetID, _ := strconv.ParseInt(os.Getenv("PROGRESS_SHEET_ID"), 10, 64)

	var opts []sevent.HandlerOptsFunc

	if progressSheetID != 0 {
		opts = append(opts, sevent.WithProgressSheetID(progressSheetID))
	}

	h, err := sevent.New(ctx, s, os.Getenv("SPREADSHEET_ID"), opts...)
	if err != nil {
		sugar.Fatal(err)
	}

	return h
}

func mustInitErrorEventHandler(ctx context.Context, s *sheets.Service) pubsub.Handler {
	errorSheetID, _ := strconv.ParseInt(os.Getenv("ERROR_SHEET_ID"), 10, 64)

	var opts []eevent.HandlerOptsFunc

	if errorSheetID != 0 {
		opts = append(opts, eevent.WithErrorSheetID(errorSheetID))
	}

	h, err := eevent.NewHandler(ctx, s, os.Getenv("SPREADSHEET_ID"), opts...)
	if err != nil {
		sugar.Fatal(err)
	}

	return h
}

func mustSetTimezone(tz string) {
	loc, err := time.LoadLocation(tz)
	if err != nil {
		sugar.Fatal(err)
	}

	time.Local = loc
}

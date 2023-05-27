package main

import (
	"context"
	"log"
	"os"
	"strconv"
	"time"

	_ "github.com/joho/godotenv/autoload"
	"github.com/piatoss3612/presentation-helper-bot/internal/config"
	"github.com/piatoss3612/presentation-helper-bot/internal/event/msgqueue"
	"github.com/piatoss3612/presentation-helper-bot/internal/logger/app"
	"github.com/piatoss3612/presentation-helper-bot/internal/logger/service"
	"github.com/piatoss3612/presentation-helper-bot/internal/tools"
	"go.uber.org/zap"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

var sugar *zap.SugaredLogger

func main() {
	l, _ := zap.NewProduction()
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

	svc := mustInitLoggerService()

	sugar.Info("Logger service is ready!")

	logger := app.New(svc, sub, sugar)
	stop := logger.Run()

	logger.Listen(stop, cfg.RabbitMQ.Topics)
}

func mustLoadConfig(path string) *config.LoggerConfig {
	cfg, err := config.NewLoggerConfig(path)
	if err != nil {
		sugar.Fatal(err)
	}

	return cfg
}

func mustInitSubscriber(ctx context.Context, addr, exchange, kind, queue string) (msgqueue.Subscriber, func() error) {
	rabbit := <-tools.RedialRabbitMQ(ctx, addr)

	if rabbit == nil {
		sugar.Fatal("Failed to connect to RabbitMQ")
	}

	sub, err := msgqueue.NewSubscriber(rabbit, exchange, kind, queue)
	if err != nil {
		log.Println(err)
		sugar.Fatal(err)
	}

	return sub, func() error { return rabbit.Close() }
}

func mustInitLoggerService() service.Service {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	eventSheetID, err := strconv.ParseInt(os.Getenv("EVENT_SHEET_ID"), 10, 64)
	if err != nil {
		sugar.Fatal(err)
	}

	srv, err := service.New(ctx, mustInitSheetsService(), os.Getenv("SPREADSHEET_ID"), eventSheetID)
	if err != nil {
		sugar.Fatal(err)
	}

	return srv
}

func mustInitSheetsService() *sheets.Service {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

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

func mustSetTimezone(tz string) {
	loc, err := time.LoadLocation(tz)
	if err != nil {
		sugar.Fatal(err)
	}

	time.Local = loc
}

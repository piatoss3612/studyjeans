package main

import (
	"context"
	"os"
	"time"

	_ "github.com/joho/godotenv/autoload"
	app "github.com/piatoss3612/presentation-helper-bot/internal/app/recorder"
	"github.com/piatoss3612/presentation-helper-bot/internal/service/recorder"
	"go.uber.org/zap"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
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
	srv := mustInitRecorderService()

	sugar.Info("Recorder service is ready!")

	rest := app.New(srv, sugar)
	<-rest.Run()
}

func mustInitRecorderService() recorder.Service {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	srv, err := recorder.New(ctx, mustInitSheetsService(), os.Getenv("SPREADSHEET_ID"), 0)
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

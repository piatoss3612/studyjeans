package app

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/piatoss3612/my-study-bot/internal/logger/service"
	"github.com/piatoss3612/my-study-bot/internal/msgqueue"
	"go.uber.org/zap"
)

type LoggerApp struct {
	svc service.Service
	sub msgqueue.Subscriber

	sugar *zap.SugaredLogger
}

func New(svc service.Service, sub msgqueue.Subscriber, sugar *zap.SugaredLogger) *LoggerApp {
	return &LoggerApp{
		svc:   svc,
		sub:   sub,
		sugar: sugar,
	}
}

func (l *LoggerApp) Run() <-chan bool {
	l.sugar.Info("Starting logger app")

	stop := make(chan bool)
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

		l.sugar.Info("Shutting down recorder server")
	}()

	return stop
}

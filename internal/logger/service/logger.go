package service

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/piatoss3612/my-study-bot/internal/pubsub"
	"go.uber.org/zap"
)

type LoggerService struct {
	sub    pubsub.Subscriber
	mapper pubsub.Mapper

	sugar *zap.SugaredLogger
}

func New(sub pubsub.Subscriber, mapper pubsub.Mapper, sugar *zap.SugaredLogger) *LoggerService {
	return &LoggerService{
		sub:    sub,
		mapper: mapper,
		sugar:  sugar,
	}
}

func (l *LoggerService) Run() <-chan bool {
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
	}()

	return stop
}

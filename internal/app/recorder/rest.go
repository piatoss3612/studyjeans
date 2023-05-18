package recorder

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/piatoss3612/presentation-helper-bot/internal/msgqueue"
	"github.com/piatoss3612/presentation-helper-bot/internal/service/recorder"
	"go.uber.org/zap"
)

type Recorder struct {
	svc recorder.Service
	sub msgqueue.Subscriber

	sugar *zap.SugaredLogger
}

func New(svc recorder.Service, sub msgqueue.Subscriber, sugar *zap.SugaredLogger) *Recorder {
	return &Recorder{
		svc:   svc,
		sub:   sub,
		sugar: sugar,
	}
}

func (r *Recorder) Run() <-chan bool {
	r.sugar.Info("Starting recorder")

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

		r.sugar.Info("Shutting down recorder server")
	}()

	return stop
}

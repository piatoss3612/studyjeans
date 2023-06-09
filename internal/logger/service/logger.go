package service

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/piatoss3612/my-study-bot/internal/pubsub"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

var metricServerPort = "8080"

type LoggerService struct {
	sub    pubsub.Subscriber
	mapper pubsub.Mapper

	srv *http.Server

	sugar *zap.SugaredLogger
}

func New(sub pubsub.Subscriber, mapper pubsub.Mapper, sugar *zap.SugaredLogger) *LoggerService {
	svc := &LoggerService{
		sub:    sub,
		mapper: mapper,
		sugar:  sugar,
	}
	return svc.setup()
}

func (l *LoggerService) setup() *LoggerService {
	metrics := prometheus.NewRegistry()
	metrics.MustRegister(collectors.NewGoCollector())

	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.HandlerFor(metrics, promhttp.HandlerOpts{}))
	mux.HandleFunc("/healthcheck", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	l.srv = &http.Server{
		Addr:    fmt.Sprintf(":%s", metricServerPort),
		Handler: mux,
	}

	return l
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
		if err := l.srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			l.sugar.Fatal("Failed to start metric server", "error", err)
		}
	}()

	go func() {
		defer func() {
			close(shutdown)
			close(stop)
		}()
		<-shutdown
	}()

	return stop
}

func (l *LoggerService) Close(ctx context.Context) error {
	if err := l.srv.Shutdown(ctx); err != nil {
		return err
	}
	return nil
}

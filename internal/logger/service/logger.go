package service

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/piatoss3612/my-study-bot/internal/pubsub"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

var totalEvents = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "listener_events_total",
		Help: "Total number of events.",
	},
	[]string{"event"},
)

var totalErrors = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "listener_errors_total",
		Help: "Total number of errors.",
	},
	[]string{"event"},
)

var duration = prometheus.NewHistogramVec(
	prometheus.HistogramOpts{
		Name: "listener_response_time_seconds",
		Help: "Response time of listener.",
	},
	[]string{"event"},
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
	metrics.MustRegister(totalEvents)
	metrics.MustRegister(totalErrors)
	metrics.MustRegister(duration)

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

func (l *LoggerService) Listen(stop <-chan bool, topics []string) {
	msgs, errs, close, err := l.sub.Subscribe(topics...)
	if err != nil {
		l.sugar.Fatal(err)
	}
	defer close()

	for {
		select {
		case msg := <-msgs:
			h, ok := l.mapper.Map(msg.Topic)
			if !ok {
				l.sugar.Errorw("Unknown event name", "event", msg.Topic)
				continue
			}

			totalEvents.WithLabelValues(msg.Topic).Inc()

			timer := prometheus.NewTimer(duration.WithLabelValues(msg.Topic))

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			if err := h.Handle(ctx, msg.Body); err != nil {
				totalErrors.WithLabelValues(msg.Topic).Inc()
				l.sugar.Errorw("Failed to handle event", "event", msg.Topic, "error", err, "duration", timer.ObserveDuration().String())
				continue
			}

			l.sugar.Infow("Successfully handled event", "event", msg.Topic, "duration", timer.ObserveDuration().String())
		case err := <-errs:
			if err == nil {
				continue
			}
			l.sugar.Errorw("Received error from subscriber", "error", err)
		case <-stop:
			l.sugar.Info("Stop listening to events")
			return
		default:
			time.Sleep(500 * time.Millisecond)
		}
	}
}

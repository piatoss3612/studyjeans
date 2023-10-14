package metrics

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	ErrorInvalidPort    = errors.New("invalid port")
	ErrorInvalidWriteTo = errors.New("invalid write timeout")
	ErrorInvalidReadTo  = errors.New("invalid read timeout")
)

type ServerConfig struct {
	Port         string
	WriteTimeout time.Duration
	ReadTimeout  time.Duration
}

func NewHttpMetricsServer(reg prometheus.Gatherer, opts promhttp.HandlerOpts, cfg ServerConfig, middlewares ...func(http.Handler) http.Handler) (*http.Server, error) {
	numPort, err := strconv.Atoi(cfg.Port)
	if err != nil {
		return nil, err
	}

	if numPort < 0 || numPort > 65535 {
		return nil, ErrorInvalidPort
	}

	if cfg.WriteTimeout <= 0 {
		return nil, ErrorInvalidWriteTo
	}

	if cfg.ReadTimeout <= 0 {
		return nil, ErrorInvalidReadTo
	}

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	for _, m := range middlewares {
		r.Use(m)
	}

	r.Handle("/metrics", promhttp.HandlerFor(reg, opts))
	r.Handle("/healthcheck", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	return &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      r,
		WriteTimeout: cfg.WriteTimeout,
		ReadTimeout:  cfg.ReadTimeout,
	}, nil
}

package bot

import (
	"net/http"
	"time"

	"github.com/piatoss3612/studyjeans/pkg/metrics"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	totalRequests = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "command_requests_total",
			Help: "Total number of requests.",
		},
		[]string{"application_command"},
	)
	totalSuccess = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "command_success_total",
			Help: "Total number of successful requests.",
		},
		[]string{"application_command"},
	)
	totalErrors = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "command_errors_total",
			Help: "Total number of errors.",
		},
		[]string{"application_command"},
	)
	duration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "command_response_duration_seconds",
			Help:    "Command response duration distribution.",
			Buckets: []float64{0.1, 0.5, 1, 2, 5, 10, 30, 60},
		},
		[]string{"application_command"},
	)
	totalGuilds = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "guilds_total",
			Help: "Total number of guilds.",
		},
	)
)

func NewBotMetricsServer(port string) (*http.Server, error) {
	reg := prometheus.NewRegistry()
	reg.MustRegister(
		collectors.NewGoCollector(),
		totalRequests,
		totalSuccess,
		totalErrors,
		duration,
		totalGuilds,
	)

	return metrics.NewHttpMetricsServer(reg, promhttp.HandlerOpts{}, metrics.ServerConfig{
		Port:         port,
		WriteTimeout: 10 * time.Second,
		ReadTimeout:  10 * time.Second,
	})
}

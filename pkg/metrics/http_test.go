package metrics

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func TestNewHttpMetricsServer(t *testing.T) {
	tests := []struct {
		name   string
		cfg    ServerConfig
		mwares []func(http.Handler) http.Handler
		hasErr bool
		expErr string
	}{
		{
			name: "valid config",
			cfg: ServerConfig{
				Port:         "8080",
				WriteTimeout: 1,
				ReadTimeout:  1,
			},
			mwares: []func(http.Handler) http.Handler{func(h http.Handler) http.Handler { return h }},
			hasErr: false,
			expErr: "",
		},
		{
			name: "invalid port exceeds max",
			cfg: ServerConfig{
				Port:         "65536",
				WriteTimeout: 1,
				ReadTimeout:  1,
			},
			hasErr: true,
			expErr: ErrorInvalidPort.Error(),
		},
		{
			name: "invalid port is not a number",
			cfg: ServerConfig{
				Port:         "abc",
				WriteTimeout: 1,
				ReadTimeout:  1,
			},
			hasErr: true,
			expErr: "strconv.Atoi: parsing \"abc\": invalid syntax",
		},
		{
			name: "invalid write timeout",
			cfg: ServerConfig{
				Port:         "8080",
				WriteTimeout: 0,
				ReadTimeout:  1,
			},
			hasErr: true,
			expErr: ErrorInvalidWriteTo.Error(),
		},
		{
			name: "invalid read timeout",
			cfg: ServerConfig{
				Port:         "8080",
				WriteTimeout: 1,
				ReadTimeout:  0,
			},
			hasErr: true,
			expErr: ErrorInvalidReadTo.Error(),
		},
	}

	reg := prometheus.NewRegistry()
	opts := promhttp.HandlerOpts{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewHttpMetricsServer(reg, opts, tt.cfg, tt.mwares...)
			if tt.hasErr && !strings.Contains(err.Error(), tt.expErr) {
				t.Errorf("expected error to be '%v', got '%v'", tt.expErr, err)
			}
		})
	}
}

func TestHttpMetricsServerHealthcheck(t *testing.T) {
	reg := prometheus.NewRegistry()
	opts := promhttp.HandlerOpts{}
	cfg := ServerConfig{
		Port:         "8080",
		WriteTimeout: 1,
		ReadTimeout:  1,
	}

	srv, err := NewHttpMetricsServer(reg, opts, cfg)
	if err != nil {
		t.Fatal(err)
	}

	tsrv := httptest.NewServer(srv.Handler)

	r := httptest.NewRecorder()
	req, err := http.NewRequest(http.MethodGet, tsrv.URL+"/healthcheck", nil)
	if err != nil {
		t.Fatal(err)
	}

	srv.Handler.ServeHTTP(r, req)

	if r.Code != http.StatusOK {
		t.Errorf("expected status code to be %v, got %v", http.StatusOK, r.Code)
	}

	tsrv.Close()
}

func TestHttpMetricsServerMetrics(t *testing.T) {
	reg := prometheus.NewRegistry()
	counter := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "test_counter",
		Help: "test counter",
	})
	reg.MustRegister(counter)
	opts := promhttp.HandlerOpts{}
	cfg := ServerConfig{
		Port:         "8080",
		WriteTimeout: 1,
		ReadTimeout:  1,
	}

	srv, err := NewHttpMetricsServer(reg, opts, cfg)
	if err != nil {
		t.Fatal(err)
	}

	tsrv := httptest.NewServer(srv.Handler)

	counter.Inc()

	r := httptest.NewRecorder()
	req, err := http.NewRequest(http.MethodGet, tsrv.URL+"/metrics", nil)
	if err != nil {
		t.Fatal(err)
	}

	srv.Handler.ServeHTTP(r, req)

	if r.Code != http.StatusOK {
		t.Errorf("expected status code to be %v, got %v", http.StatusOK, r.Code)
	}

	if !strings.Contains(r.Body.String(), "test_counter 1") {
		t.Errorf("expected body to contain 'test_counter 1', got '%v'", r.Body.String())
	}

	tsrv.Close()
}

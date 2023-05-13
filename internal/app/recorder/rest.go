package recorder

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/piatoss3612/presentation-helper-bot/internal/service/recorder"
	"go.uber.org/zap"
	"golang.org/x/net/context"
)

var defaultPort = "8080"

type Rest struct {
	*http.Server

	svc recorder.Service
	// TODO: message consumer

	sugar *zap.SugaredLogger
}

func New(svc recorder.Service, sugar *zap.SugaredLogger, port ...string) *Rest {
	srv := &http.Server{
		Addr: fmt.Sprintf(":%s", defaultPort),
	}

	if len(port) > 0 {
		srv.Addr = fmt.Sprintf(":%s", port[0])
	}

	return &Rest{
		Server: srv,
		svc:    svc,
		sugar:  sugar,
	}
}

func (r *Rest) Run() <-chan bool {
	r.sugar.Info("Starting recorder server on port", r.Addr)

	stop := make(chan bool)
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT,
	)

	go func() {
		if err := r.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			r.sugar.Fatalw(fmt.Sprintf("Could not listen on %s", r.Addr), "error", err)
		}
	}()

	go func() {
		defer func() {
			close(shutdown)
			close(stop)
		}()
		<-shutdown

		r.sugar.Info("Shutting down recorder server")

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := r.Shutdown(ctx); err != nil {
			r.sugar.Fatalf("Could not gracefully shutdown the server: %v", err)
		}
	}()

	return stop
}

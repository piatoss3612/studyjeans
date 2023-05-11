package recorder

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"go.uber.org/zap"
	"golang.org/x/net/context"
	"google.golang.org/api/sheets/v4"
)

type Rest struct {
	*http.Server

	sheetsSrv     *sheets.Service
	spreadsheetID string

	// TODO: message consumer

	sugar *zap.SugaredLogger
}

func New(sheetsSrv *sheets.Service, spreadsheetID string, port string, sugar *zap.SugaredLogger) *Rest {
	_, err := strconv.Atoi(port)
	if err != nil {
		port = "8080"
	}

	srv := &http.Server{
		Addr: fmt.Sprintf(":%s", port),
	}

	return &Rest{
		Server:        srv,
		sheetsSrv:     sheetsSrv,
		spreadsheetID: spreadsheetID,
		sugar:         sugar,
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

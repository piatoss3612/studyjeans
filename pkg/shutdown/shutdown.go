package shutdown

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
)

func GracefulShutdown(fn func() error, sigs ...os.Signal) <-chan struct{} {
	stop := make(chan struct{})
	sigChan := make(chan os.Signal, 1)

	signal.Notify(sigChan, sigs...)

	go func() {
		defer func() {
			if r := recover(); r != nil {
				slog.Info("Panic recovered in graceful shutdown", slog.Any("panic", r))
			}
		}()

		<-sigChan
		if err := fn(); err != nil {
			panic(err)
		}

		close(stop)
	}()

	return stop
}

func GracefulShutdownCtx(ctx context.Context, fn func() error, sigs ...os.Signal) (context.Context, context.CancelFunc) {
	ctx, cancel := signal.NotifyContext(ctx, sigs...)

	go func() {
		defer func() {
			if r := recover(); r != nil {
				slog.Info("Panic recovered in graceful shutdown", slog.Any("panic", r))
			}
		}()

		<-ctx.Done()
		if err := fn(); err != nil {
			panic(err)
		}
	}()

	return ctx, cancel
}

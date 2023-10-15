package shutdown

import (
	"context"
	"os"
	"syscall"
	"testing"
)

func TestGracefulShutdown(t *testing.T) {
	fn := func() {}

	stop := GracefulShutdown(fn, os.Interrupt)
	if stop == nil {
		t.Error("stop channel is nil")
	}

	err := syscall.Kill(syscall.Getpid(), syscall.SIGINT)
	if err != nil {
		t.Error(err)
	}

	_, ok := <-stop
	if ok {
		t.Error("stop channel is not closed")
	}
}

func TestGracefulShutdownCtx(t *testing.T) {
	fn := func() {}

	ctx, cancel := GracefulShutdownCtx(context.Background(), fn, os.Interrupt)
	if ctx == nil {
		t.Error("ctx is nil")
	}

	if cancel == nil {
		t.Error("cancel func is nil")
	}

	err := syscall.Kill(syscall.Getpid(), syscall.SIGINT)
	if err != nil {
		t.Error(err)
	}

	_, ok := <-ctx.Done()
	if ok {
		t.Error("ctx is not closed")
	}

	if ctx.Err() != context.Canceled {
		t.Error("ctx error is not context.Canceled")
	}
}

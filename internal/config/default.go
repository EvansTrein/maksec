package config

import (
	"context"
	"os/signal"
	"syscall"
)

func DefaultCtx() context.Context {
	return context.Background()
}

func DefaultCtxRootSysNotify(ctx context.Context) (context.Context, context.CancelFunc) {
	ctx, cancel := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	return ctx, cancel
}

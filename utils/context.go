package utils

import (
	"context"
	"time"
)

const defaultTimeout = 5

func DefaultCtx() (ctx context.Context, cancel context.CancelFunc) {
	ctx, cancel = context.WithTimeout(context.Background(), defaultTimeout*time.Second)
	return
}

func MakeCtx(duration time.Duration) (ctx context.Context, cancel context.CancelFunc) {
	ctx, cancel = context.WithTimeout(context.Background(), duration)
	return
}

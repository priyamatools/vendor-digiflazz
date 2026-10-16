package database

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	defaultMaxAttempts = 30
	defaultRetryDelay    = 2 * time.Second
)

// ConnectPool menunggu PostgreSQL siap (retry saat startup / docker restart).
func ConnectPool(ctx context.Context, databaseURL string, log *slog.Logger) (*pgxpool.Pool, error) {
	var lastErr error
	for attempt := 1; attempt <= defaultMaxAttempts; attempt++ {
		pool, err := pgxpool.New(ctx, databaseURL)
		if err != nil {
			lastErr = err
			log.Warn("database connect", "attempt", attempt, "err", err)
			if !sleepCtx(ctx, defaultRetryDelay) {
				return nil, ctx.Err()
			}
			continue
		}
		if err := pool.Ping(ctx); err != nil {
			pool.Close()
			lastErr = err
			log.Warn("database ping", "attempt", attempt, "err", err)
			if !sleepCtx(ctx, defaultRetryDelay) {
				return nil, ctx.Err()
			}
			continue
		}
		if attempt > 1 {
			log.Info("database ready", "attempt", attempt)
		}
		return pool, nil
	}
	return nil, fmt.Errorf("database unavailable after %d attempts: %w", defaultMaxAttempts, lastErr)
}

// sleepCtx menunggu durasi atau dibatalkan oleh context.
func sleepCtx(ctx context.Context, d time.Duration) bool {
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return false
	case <-t.C:
		return true
	}
}

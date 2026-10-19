package worker

import (
	"context"
	"log/slog"
	"time"

	"github.com/addon-vendor-digiflazz/internal/config"
	"github.com/addon-vendor-digiflazz/internal/service"
)

// RunPendingReconcile menjalankan loop worker poll checktoday untuk transaksi pending.
func RunPendingReconcile(ctx context.Context, cfg config.Config, svc *service.TopupService, log *slog.Logger) {
	if !cfg.PendingWorkerEnabled {
		log.Info("pending reconcile worker disabled")
		return
	}

	ticker := time.NewTicker(cfg.PendingWorkerInterval)
	defer ticker.Stop()

	log.Info("pending reconcile worker started",
		"interval", cfg.PendingWorkerInterval.String(),
		"batch", cfg.PendingWorkerBatchSize,
		"timezone", cfg.ReconcileTimezone,
	)

	for {
		select {
		case <-ctx.Done():
			log.Info("pending reconcile worker stopped")
			return
		case <-ticker.C:
			n, err := svc.ReconcilePendingBatch(ctx)
			if err != nil {
				log.Error("pending reconcile batch", "err", err)
				continue
			}
			if n > 0 {
				log.Info("pending reconcile updated", "count", n)
			}
		}
	}
}

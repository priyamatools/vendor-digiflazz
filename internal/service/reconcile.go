package service

import (
	"context"
	"strings"
	"time"

	"github.com/addon-vendor-digiflazz/internal/atrilinks"
	"github.com/addon-vendor-digiflazz/internal/digiflazz"
	"github.com/addon-vendor-digiflazz/internal/repository"
)

// ReconcilePendingBatch mem-poll checktoday untuk transaksi pending (dipanggil worker).
func (s *TopupService) ReconcilePendingBatch(ctx context.Context) (int, error) {
	list, err := s.repo.ListPendingForReconcile(
		ctx,
		s.cfg.ReconcileTimezone,
		s.cfg.PendingWorkerBatchSize,
		s.cfg.PendingReconcileMinAge,
	)
	if err != nil {
		return 0, err
	}
	var updated int
	for _, tx := range list {
		_, reconciled, err := s.reconcilePending(ctx, tx, true)
		if err != nil {
			return updated, err
		}
		if reconciled {
			updated++
		}
	}
	return updated, nil
}

// reconcilePending memanggil checktoday bila transaksi masih pending dan memenuhi syarat hari ini.
func (s *TopupService) reconcilePending(
	ctx context.Context,
	tx repository.Transaction,
	throttle bool,
) (digiflazz.ResponseData, bool, error) {
	if tx.DigiflazzStatus != "0" || !tx.OrderSubmitted {
		return digiflazz.ResponseData{}, false, nil
	}
	if !s.isCreatedToday(tx.CreatedAt) {
		return digiflazz.ResponseData{}, false, nil
	}
	if throttle {
		acquired, err := s.repo.AcquireStatusCheck(ctx, tx.RefID, s.cfg.PendingReconcileThrottle)
		if err != nil {
			return digiflazz.ResponseData{}, false, err
		}
		if !acquired {
			return digiflazz.ResponseData{}, false, nil
		}
	}

	raw, up, err := s.atrilinks.CheckToday(ctx, tx.RefID, tx.HP)
	if err != nil {
		s.log.Warn("checktoday gagal", "ref_id", tx.RefID, "err", err)
		return digiflazz.ResponseData{}, false, nil
	}
	if !shouldApplyCheckToday(up) {
		return digiflazz.ResponseData{}, false, nil
	}

	data, err := s.persistUpstream(ctx, tx.RefID, tx.HP, tx.PulsaCode, tx.Price, raw, up, true, true)
	if err != nil {
		return digiflazz.ResponseData{}, false, err
	}
	return data, true, nil
}

// isCreatedToday mengecek apakah transaksi dibuat pada tanggal kalender hari ini (timezone env).
func (s *TopupService) isCreatedToday(createdAt time.Time) bool {
	loc, err := time.LoadLocation(s.cfg.ReconcileTimezone)
	if err != nil {
		loc = time.UTC
	}
	now := time.Now().In(loc)
	created := createdAt.In(loc)
	y1, m1, d1 := now.Date()
	y2, m2, d2 := created.Date()
	return y1 == y2 && m1 == m2 && d1 == d2
}

// shouldApplyCheckToday true jika response checktoday punya status final atau pending yang valid.
func shouldApplyCheckToday(up atrilinks.TransactionResponse) bool {
	st := strings.ToLower(strings.TrimSpace(up.Status))
	switch st {
	case "success", "failed", "pending":
		return true
	default:
		return false
	}
}

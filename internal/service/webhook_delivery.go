package service

import (
	"context"
	"time"

	"github.com/addon-vendor-digiflazz/internal/digiflazz"
)

// notifyDigiflazzFinal mengirim webhook ke Digiflazz bila status final dan belum terkirim.
func (s *TopupService) notifyDigiflazzFinal(ctx context.Context, refID string, data digiflazz.ResponseData, async bool) {
	if data.Status != "1" && data.Status != "2" {
		return
	}
	if async {
		s.enqueueDigiflazzWebhook(refID, data)
		return
	}
	s.deliverDigiflazzWebhook(ctx, refID, data)
}

// enqueueDigiflazzWebhook menjalankan pengiriman webhook di goroutine terpisah.
func (s *TopupService) enqueueDigiflazzWebhook(refID string, data digiflazz.ResponseData) {
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		defer cancel()
		s.deliverDigiflazzWebhook(ctx, refID, data)
	}()
}

// deliverDigiflazzWebhook mengirim webhook dengan retry singkat; idempoten via callback_sent_at.
func (s *TopupService) deliverDigiflazzWebhook(ctx context.Context, refID string, data digiflazz.ResponseData) {
	tx, err := s.repo.GetByRefID(ctx, refID)
	if err != nil {
		s.log.Error("webhook load tx", "ref_id", refID, "err", err)
		return
	}
	if tx.CallbackSentAt != nil {
		return
	}

	var lastErr error
	for attempt := 0; attempt < 3; attempt++ {
		if attempt > 0 {
			time.Sleep(time.Duration(attempt) * 2 * time.Second)
		}
		if err := s.webhook.Send(ctx, data); err != nil {
			lastErr = err
			s.log.Error("digiflazz webhook", "ref_id", refID, "attempt", attempt+1, "err", err)
			continue
		}
		if err := s.repo.MarkCallbackSent(ctx, refID); err != nil {
			s.log.Error("mark callback sent", "ref_id", refID, "err", err)
		}
		return
	}
	if lastErr != nil {
		s.log.Error("digiflazz webhook exhausted", "ref_id", refID, "err", lastErr)
	}
}

package service

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"strings"

	"github.com/addon-vendor-digiflazz/internal/atrilinks"
	"github.com/addon-vendor-digiflazz/internal/config"
	"github.com/addon-vendor-digiflazz/internal/digiflazz"
	"github.com/addon-vendor-digiflazz/internal/repository"
)

type TopupService struct {
	cfg      config.Config
	repo     *repository.Repository
	atrilinks *atrilinks.Client
	webhook  *digiflazz.WebhookClient
	log      *slog.Logger
}

func NewTopupService(
	cfg config.Config,
	repo *repository.Repository,
	atClient *atrilinks.Client,
	webhook *digiflazz.WebhookClient,
	log *slog.Logger,
) *TopupService {
	return &TopupService{
		cfg:       cfg,
		repo:      repo,
		atrilinks: atClient,
		webhook:   webhook,
		log:       log,
	}
}

func (s *TopupService) HandleTopup(ctx context.Context, req digiflazz.TopupRequest) (digiflazz.ResponseData, error) {
	if req.Username != s.cfg.DigiflazzUsername {
		return digiflazz.WrongAuthResponse(req.RefID, req.HP, req.PulsaCode, req.Price), nil
	}
	if !digiflazz.ValidSign(req.Username, s.cfg.DigiflazzAPIKey, req.RefID, req.Sign) {
		return digiflazz.WrongAuthResponse(req.RefID, req.HP, req.PulsaCode, req.Price), nil
	}
	if strings.ToLower(strings.TrimSpace(req.Commands)) != "topup" {
		return digiflazz.ResponseData{
			RefID: req.RefID, Status: "2", Code: req.PulsaCode, HP: req.HP,
			Price: digiflazz.PriceString(req.Price), Message: "FAILED", Balance: "0",
			TrID: req.RefID, RC: "07", SN: "",
		}, nil
	}
	if req.RefID == "" || req.HP == "" || req.PulsaCode == "" {
		return digiflazz.ResponseData{
			RefID: req.RefID, Status: "2", Code: req.PulsaCode, HP: req.HP,
			Price: digiflazz.PriceString(req.Price), Message: "FAILED", Balance: "0",
			TrID: req.RefID, RC: "07", SN: "",
		}, nil
	}

	existing, err := s.repo.GetByRefID(ctx, req.RefID)
	if err == nil {
		if existing.OrderSubmitted {
			return digiflazz.FromStored(
				existing.RefID, existing.HP, existing.PulsaCode, existing.Price, existing.ResponsePrice,
				existing.DigiflazzStatus, existing.DigiflazzRC, existing.Message,
				existing.SN, existing.Balance,
			), nil
		}
	} else if !errors.Is(err, repository.ErrNotFound) {
		return digiflazz.ResponseData{}, err
	}

	if errors.Is(err, repository.ErrNotFound) {
		if err := s.repo.CreatePending(ctx, repository.Transaction{
			RefID:             req.RefID,
			HP:                req.HP,
			PulsaCode:         strings.ToLower(req.PulsaCode),
			Price:             req.Price,
			DigiflazzUsername: req.Username,
		}); err != nil {
			return digiflazz.ResponseData{}, err
		}
	}

	kodeProduk := strings.ToLower(req.PulsaCode)
	raw, up, err := s.atrilinks.Order(ctx, atrilinks.OrderParams{
		ReqID:      req.RefID,
		KodeProduk: kodeProduk,
		Tujuan:     req.HP,
		JenisTrx:   1,
		LimitHarga: req.Price,
	})
	if err != nil {
		s.log.Error("atrilinks order", "ref_id", req.RefID, "err", err)
		return digiflazz.ResponseData{
			RefID: req.RefID, Status: "2", Code: kodeProduk, HP: req.HP,
			Price: digiflazz.PriceString(req.Price), Message: "FAILED", Balance: "0",
			TrID: req.RefID, RC: "07", SN: "",
		}, nil
	}

	return s.persistUpstream(ctx, req.RefID, req.HP, kodeProduk, req.Price, raw, up, true)
}

func (s *TopupService) HandleAtrilinksCallback(ctx context.Context, raw []byte) error {
	up, err := atrilinks.ParseTransactionResponse(raw)
	if err != nil {
		return err
	}
	if up.ReqID == "" {
		return errors.New("callback tanpa reqid")
	}

	tx, err := s.repo.GetByRefID(ctx, up.ReqID)
	if errors.Is(err, repository.ErrNotFound) {
		return err
	}
	if err != nil {
		return err
	}

	_, err = s.persistUpstream(ctx, tx.RefID, tx.HP, tx.PulsaCode, tx.Price, raw, up, true)
	return err
}

func (s *TopupService) persistUpstream(
	ctx context.Context,
	refID, hp, pulsaCode string,
	price int64,
	raw []byte,
	up atrilinks.TransactionResponse,
	orderSubmitted bool,
) (digiflazz.ResponseData, error) {
	up = s.enrichUpstreamHarga(ctx, refID, up, raw)
	data := digiflazz.MapFromUpstream(up, refID, hp, pulsaCode, price)

	var rawJSON json.RawMessage
	if len(raw) > 0 {
		rawJSON = json.RawMessage(raw)
	}

	err := s.repo.ApplyUpstream(ctx, repository.UpstreamUpdate{
		RefID:           refID,
		AtrilinksStatus: up.Status,
		DigiflazzStatus: data.Status,
		DigiflazzRC:     data.RC,
		Message:         data.Message,
		SN:              data.SN,
		Balance:         data.Balance,
		ResponsePrice:   data.Price,
		UpstreamRaw:     rawJSON,
		OrderSubmitted:  orderSubmitted,
	})
	if err != nil {
		return digiflazz.ResponseData{}, err
	}

	if data.Status == "1" || data.Status == "2" {
		tx, err := s.repo.GetByRefID(ctx, refID)
		if err != nil {
			return data, err
		}
		if tx.CallbackSentAt == nil {
			if err := s.webhook.Send(ctx, data); err != nil {
				s.log.Error("digiflazz webhook", "ref_id", refID, "err", err)
			} else if err := s.repo.MarkCallbackSent(ctx, refID); err != nil {
				s.log.Error("mark callback sent", "ref_id", refID, "err", err)
			}
		}
	}

	return data, nil
}

func (s *TopupService) enrichUpstreamHarga(ctx context.Context, refID string, up atrilinks.TransactionResponse, raw []byte) atrilinks.TransactionResponse {
	if up.Harga > 0 {
		return up
	}
	if h := atrilinks.HargaFromRaw(raw); h > 0 {
		up.Harga = h
		return up
	}
	tx, err := s.repo.GetByRefID(ctx, refID)
	if err != nil {
		return up
	}
	if h := atrilinks.HargaFromRaw(tx.UpstreamRaw); h > 0 {
		up.Harga = h
	}
	return up
}

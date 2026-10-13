package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/addon-vendor-digiflazz/internal/digiflazz"
	"github.com/addon-vendor-digiflazz/internal/httplog"
	"github.com/addon-vendor-digiflazz/internal/service"
)

type DigiflazzHandler struct {
	svc *service.TopupService
	log *slog.Logger
	dev bool
}

func NewDigiflazzHandler(svc *service.TopupService, log *slog.Logger, development bool) *DigiflazzHandler {
	return &DigiflazzHandler{svc: svc, log: log, dev: development}
}

func (h *DigiflazzHandler) Topup(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req digiflazz.TopupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	if h.dev {
		httplog.DigiflazzRequest(h.log, req.Username, req.Commands, req.RefID, req.HP, req.PulsaCode, req.Price, req.Sign)
	}

	data, err := h.svc.HandleTopup(r.Context(), req)
	if err != nil {
		if h.dev {
			h.log.Debug("digiflazz handler error", "ref_id", req.RefID, "err", err.Error())
		}
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if h.dev {
		httplog.DigiflazzResponse(h.log, data.RefID, data.Status, data.RC, data.Code, data.HP, data.Price, data.Message, data.Balance, data.TrID, data.SN)
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(digiflazz.Envelope{Data: data})
}

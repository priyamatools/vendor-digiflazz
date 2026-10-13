package handler

import (
	"encoding/json"
	"net/http"

	"github.com/addon-vendor-digiflazz/internal/digiflazz"
	"github.com/addon-vendor-digiflazz/internal/service"
)

type DigiflazzHandler struct {
	svc *service.TopupService
}

func NewDigiflazzHandler(svc *service.TopupService) *DigiflazzHandler {
	return &DigiflazzHandler{svc: svc}
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

	data, err := h.svc.HandleTopup(r.Context(), req)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(digiflazz.Envelope{Data: data})
}

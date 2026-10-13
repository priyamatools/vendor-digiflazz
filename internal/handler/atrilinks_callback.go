package handler

import (
	"io"
	"net/http"

	"github.com/addon-vendor-digiflazz/internal/service"
)

type AtrilinksCallbackHandler struct {
	svc *service.TopupService
}

func NewAtrilinksCallbackHandler(svc *service.TopupService) *AtrilinksCallbackHandler {
	return &AtrilinksCallbackHandler{svc: svc}
}

func (h *AtrilinksCallbackHandler) Callback(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost && r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	raw, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		http.Error(w, "bad body", http.StatusBadRequest)
		return
	}
	if len(raw) == 0 && r.Method == http.MethodGet {
		raw = []byte(r.URL.Query().Get("payload"))
	}
	if len(raw) == 0 {
		http.Error(w, "empty body", http.StatusBadRequest)
		return
	}

	if err := h.svc.HandleAtrilinksCallback(r.Context(), raw); err != nil {
		http.Error(w, "callback failed", http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}

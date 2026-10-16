package handler

import (
	"log/slog"
	"net/http"

	"github.com/addon-vendor-digiflazz/internal/atrilinks"
	"github.com/addon-vendor-digiflazz/internal/httplog"
	"github.com/addon-vendor-digiflazz/internal/service"
)

// AtrilinksCallbackHandler menangani urlcallback dari Atrilinks (GET/POST).
type AtrilinksCallbackHandler struct {
	svc *service.TopupService
	log *slog.Logger
	dev bool
}

// NewAtrilinksCallbackHandler membuat handler callback upstream.
func NewAtrilinksCallbackHandler(svc *service.TopupService, log *slog.Logger, development bool) *AtrilinksCallbackHandler {
	return &AtrilinksCallbackHandler{svc: svc, log: log, dev: development}
}

// Callback mem-parse payload (termasuk query msg) dan memperbarui transaksi.
func (h *AtrilinksCallbackHandler) Callback(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet, http.MethodPost, http.MethodPut, http.MethodPatch:
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if h.dev {
		h.log.Debug("atrilinks callback hit",
			"method", r.Method,
			"content_type", r.Header.Get("Content-Type"),
			"query", r.URL.RawQuery,
		)
	}

	raw, err := atrilinks.ExtractCallbackPayload(r)
	if err != nil {
		if h.dev {
			h.log.Debug("atrilinks callback parse payload", "err", err.Error())
		}
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	if h.dev {
		httplog.AtrilinksCallback(h.log, raw)
	}

	if err := h.svc.HandleAtrilinksCallback(r.Context(), raw); err != nil {
		if h.dev {
			h.log.Debug("atrilinks callback handle", "err", err.Error())
		}
		http.Error(w, "callback failed", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}

package httplog

import (
	"log/slog"
	"net/url"
)

// RedactAtrilinksURL masks pin/pass in query string for logs.
func RedactAtrilinksURL(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return "[invalid-url]"
	}
	q := u.Query()
	if q.Has("pin") {
		q.Set("pin", "***")
	}
	if q.Has("pass") {
		q.Set("pass", "***")
	}
	u.RawQuery = q.Encode()
	return u.String()
}

func DigiflazzRequest(log *slog.Logger, username, commands, refID, hp, pulsaCode string, price int64, sign string) {
	log.Debug("digiflazz request",
		"username", username,
		"commands", commands,
		"ref_id", refID,
		"hp", hp,
		"pulsa_code", pulsaCode,
		"price", price,
		"sign", sign,
	)
}

func DigiflazzResponse(log *slog.Logger, refID, status, rc, code, hp, price, message, balance, trID, sn string) {
	log.Debug("digiflazz response",
		"ref_id", refID,
		"status", status,
		"rc", rc,
		"code", code,
		"hp", hp,
		"price", price,
		"message", message,
		"balance", balance,
		"tr_id", trID,
		"sn", sn,
	)
}

func AtrilinksOrder(log *slog.Logger, endpoint, safeURL, reqid, kodeproduk, tujuan string, jenistrx int, limitharga int64) {
	log.Debug("atrilinks request",
		"endpoint", endpoint,
		"url", safeURL,
		"reqid", reqid,
		"kodeproduk", kodeproduk,
		"tujuan", tujuan,
		"jenistrx", jenistrx,
		"limitharga", limitharga,
	)
}

func AtrilinksResponse(log *slog.Logger, refID string, httpStatus int, raw []byte, parsed map[string]string, err error) {
	if err != nil {
		log.Debug("atrilinks response error",
			"ref_id", refID,
			"http_status", httpStatus,
			"body", truncate(string(raw), 4096),
			"err", err.Error(),
		)
		return
	}
	args := []any{
		"ref_id", refID,
		"http_status", httpStatus,
		"body", truncate(string(raw), 4096),
	}
	for k, v := range parsed {
		args = append(args, k, v)
	}
	log.Debug("atrilinks response", args...)
}

func AtrilinksCallback(log *slog.Logger, body []byte) {
	log.Debug("atrilinks callback",
		"body", truncate(string(body), 4096),
	)
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "...(truncated)"
}

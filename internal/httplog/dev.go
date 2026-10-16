package httplog

import (
	"log/slog"
	"net/url"
)

// RedactAtrilinksURL menyamarkan pin, pass, dan sign pada URL untuk log.
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
	if q.Has("sign") {
		q.Set("sign", "***")
	}
	u.RawQuery = q.Encode()
	return u.String()
}

// DigiflazzRequest mencatat request masuk Digiflazz (mode development).
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

// DigiflazzResponse mencatat field response ke Digiflazz (mode development).
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

// AtrilinksOrder mencatat request order ke Atrilinks (URL sudah disamarkan).
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

// AtrilinksResponse mencatat response HTTP dari Atrilinks.
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

// AtrilinksCallback mencatat body callback dari Atrilinks.
func AtrilinksCallback(log *slog.Logger, body []byte) {
	log.Debug("atrilinks callback",
		"body", truncate(string(body), 4096),
	)
}

// truncate memotong string log yang terlalu panjang.
func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "...(truncated)"
}

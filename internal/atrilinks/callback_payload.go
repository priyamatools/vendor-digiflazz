package atrilinks

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// ExtractCallbackPayload membaca body callback Atrilinks (GET msg, POST JSON, form).
func ExtractCallbackPayload(r *http.Request) ([]byte, error) {
	if r.Method == http.MethodGet {
		return payloadFromQuery(r.URL.Query())
	}

	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		return nil, err
	}
	body = bytes.TrimSpace(body)

	if len(body) == 0 {
		return payloadFromQuery(r.URL.Query())
	}

	if body[0] == '{' || body[0] == '[' {
		return body, nil
	}

	ct := strings.ToLower(r.Header.Get("Content-Type"))
	if strings.Contains(ct, "application/x-www-form-urlencoded") || looksLikeForm(body) {
		vals, err := url.ParseQuery(string(body))
		if err != nil {
			return nil, err
		}
		if raw, ok := jsonFromMsgParam(vals); ok {
			return raw, nil
		}
		if hasCallbackFields(vals) {
			return valuesToJSON(vals)
		}
	}

	// Raw JSON without leading brace (unlikely) or single query-style string in body
	if vals, err := url.ParseQuery(string(body)); err == nil && hasCallbackFields(vals) {
		return valuesToJSON(vals)
	}

	return body, nil
}

// payloadFromQuery mengekstrak JSON callback dari query string (msg, payload, data, atau field transaksi).
func payloadFromQuery(q url.Values) ([]byte, error) {
	if p := strings.TrimSpace(q.Get("payload")); p != "" {
		return []byte(p), nil
	}
	if d := strings.TrimSpace(q.Get("data")); d != "" {
		return []byte(d), nil
	}
	if raw, ok := jsonFromMsgParam(q); ok {
		return raw, nil
	}
	if hasCallbackFields(q) {
		return valuesToJSON(q)
	}
	return nil, fmt.Errorf("callback kosong: tidak ada query/body yang dikenali")
}

// jsonFromMsgParam membaca parameter msg (GET callback: ?msg=<JSON URL-encoded>).
func jsonFromMsgParam(q url.Values) ([]byte, bool) {
	m := strings.TrimSpace(q.Get("msg"))
	if m == "" {
		return nil, false
	}
	return []byte(m), true
}

// hasCallbackFields mengecek apakah query berisi field khas notifikasi transaksi.
func hasCallbackFields(v url.Values) bool {
	if strings.TrimSpace(v.Get("reqid")) != "" {
		return true
	}
	if strings.TrimSpace(v.Get("status")) != "" {
		return true
	}
	return false
}

// valuesToJSON mengubah form/query menjadi objek JSON untuk di-parse seperti response transaksi.
func valuesToJSON(v url.Values) ([]byte, error) {
	m := make(map[string]string, len(v))
	for key, vals := range v {
		if len(vals) == 0 {
			continue
		}
		m[key] = vals[0]
	}
	return json.Marshal(m)
}

// looksLikeForm mendeteksi body berformat application/x-www-form-urlencoded.
func looksLikeForm(body []byte) bool {
	s := string(body)
	return strings.Contains(s, "reqid=") || strings.Contains(s, "status=") || strings.Contains(s, "msg=")
}

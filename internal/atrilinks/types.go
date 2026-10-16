package atrilinks

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

type TransactionResponse struct {
	ReqID  string `json:"reqid"`
	Status string `json:"status"`
	Jenis  string `json:"jenis"`
	Kode   string `json:"kode"`
	Tujuan string `json:"tujuan"`
	Harga  int64  `json:"harga"`
	Saldo  int64  `json:"saldo"`
	Bukti  string `json:"bukti"`
	Info   string `json:"info"`
	Detail string `json:"detail"`
	UserID string `json:"userid"`
}

// ParseTransactionResponse mem-parse satu baris JSON response/callback Atrilinks.
func ParseTransactionResponse(raw []byte) (TransactionResponse, error) {
	raw = bytesTrim(raw)
	if len(raw) == 0 {
		return TransactionResponse{}, fmt.Errorf("payload kosong")
	}

	var tr TransactionResponse
	if err := json.Unmarshal(raw, &tr); err == nil && tr.ReqID != "" {
		return tr, nil
	}

	var m map[string]json.RawMessage
	if err := json.Unmarshal(raw, &m); err != nil {
		return TransactionResponse{}, err
	}

	tr.ReqID = flexString(m, "reqid")
	tr.Status = flexString(m, "status")
	tr.Jenis = flexString(m, "jenis")
	tr.Kode = flexString(m, "kode")
	tr.Tujuan = flexString(m, "tujuan")
	tr.Bukti = flexString(m, "bukti")
	tr.Info = flexString(m, "info")
	tr.Detail = flexString(m, "detail")
	tr.UserID = flexString(m, "userid")
	tr.Harga, _ = flexInt64(m, "harga")
	tr.Saldo, _ = flexInt64(m, "saldo")

	if tr.ReqID == "" {
		return TransactionResponse{}, fmt.Errorf("reqid tidak ditemukan di payload")
	}
	return tr, nil
}

// bytesTrim memangkas spasi di awal/akhir byte JSON mentah.
func bytesTrim(b []byte) []byte {
	return []byte(strings.TrimSpace(string(b)))
}

// flexString membaca field string dari JSON yang bisa berupa string atau angka.
func flexString(m map[string]json.RawMessage, key string) string {
	raw, ok := m[key]
	if !ok {
		return ""
	}
	var s string
	if err := json.Unmarshal(raw, &s); err == nil {
		return s
	}
	return strings.Trim(string(raw), `"`)
}

// flexInt64 membaca field numerik dari JSON string atau number.
func flexInt64(m map[string]json.RawMessage, key string) (int64, error) {
	raw, ok := m[key]
	if !ok {
		return 0, nil
	}
	var n int64
	if err := json.Unmarshal(raw, &n); err == nil {
		return n, nil
	}
	var s string
	if err := json.Unmarshal(raw, &s); err != nil {
		return 0, err
	}
	return strconv.ParseInt(strings.TrimSpace(s), 10, 64)
}

// IsFinalStatus true jika status upstream sudah sukses atau gagal.
func IsFinalStatus(status string) bool {
	switch status {
	case "success", "failed":
		return true
	default:
		return false
	}
}

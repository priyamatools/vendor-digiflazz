package atrilinks

import "encoding/json"

// HargaFromRaw mengambil field harga dari JSON mentah upstream.
func HargaFromRaw(raw []byte) int64 {
	if len(raw) == 0 {
		return 0
	}
	var m map[string]json.RawMessage
	if err := json.Unmarshal(raw, &m); err != nil {
		return 0
	}
	h, _ := flexInt64(m, "harga")
	return h
}

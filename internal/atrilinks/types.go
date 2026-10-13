package atrilinks

import "encoding/json"

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

func ParseTransactionResponse(raw []byte) (TransactionResponse, error) {
	var tr TransactionResponse
	if err := json.Unmarshal(raw, &tr); err != nil {
		return TransactionResponse{}, err
	}
	return tr, nil
}

func IsFinalStatus(status string) bool {
	switch status {
	case "success", "failed":
		return true
	default:
		return false
	}
}

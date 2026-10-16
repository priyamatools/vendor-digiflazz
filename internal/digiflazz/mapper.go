package digiflazz

import (
	"fmt"
	"strings"

	"github.com/addon-vendor-digiflazz/internal/atrilinks"
)

// MapFromUpstream mengubah response Atrilinks menjadi field response Digiflazz.
func MapFromUpstream(up atrilinks.TransactionResponse, refID, hp, pulsaCode string, fallbackRequestPrice int64) ResponseData {
	status, rc, msg := mapStatus(up.Status, up.Info)

	price := PriceString(fallbackRequestPrice)
	if up.Harga > 0 {
		price = fmt.Sprintf("%d", up.Harga)
	}
	balance := "0"
	if up.Saldo > 0 {
		balance = fmt.Sprintf("%d", up.Saldo)
	}
	code := pulsaCode
	if up.Kode != "" {
		code = up.Kode
	}

	return ResponseData{
		RefID:   refID,
		Status:  status,
		Code:    code,
		HP:      hp,
		Price:   price,
		Message: msg,
		Balance: balance,
		TrID:    refID,
		RC:      rc,
		SN:      up.Bukti,
	}
}

// mapStatus menentukan status, rc, dan message format Digiflazz dari status kata Atrilinks.
func mapStatus(atrilinksStatus, info string) (status, rc, message string) {
	s := strings.ToLower(strings.TrimSpace(atrilinksStatus))
	switch s {
	case "success":
		msg := strings.TrimSpace(info)
		if msg == "" {
			msg = "Success"
		}
		return "1", "00", msg
	case "pending":
		msg := strings.TrimSpace(info)
		if msg == "" {
			msg = "Process"
		}
		return "0", "39", msg
	default:
		msg := strings.TrimSpace(info)
		if msg == "" {
			msg = "FAILED"
		}
		return "2", "07", msg
	}
}

// WrongAuthResponse response standar gagal autentikasi Digiflazz (rc 204).
func WrongAuthResponse(refID, hp, code string, price int64) ResponseData {
	return ResponseData{
		RefID:   refID,
		Status:  "2",
		Code:    code,
		HP:      hp,
		Price:   PriceString(price),
		Message: "WRONG AUTHENTICATION",
		Balance: "0",
		TrID:    refID,
		RC:      "204",
		SN:      "",
	}
}

// FromStored membangun response Digiflazz dari data yang tersimpan di database.
func FromStored(
	refID, hp, pulsaCode string,
	requestPrice int64,
	responsePrice string,
	digiflazzStatus, digiflazzRC, message, sn, balance string,
) ResponseData {
	price := responsePrice
	if price == "" {
		price = PriceString(requestPrice)
	}
	return ResponseData{
		RefID:   refID,
		Status:  digiflazzStatus,
		Code:    pulsaCode,
		HP:      hp,
		Price:   price,
		Message: message,
		Balance: balance,
		TrID:    refID,
		RC:      digiflazzRC,
		SN:      sn,
	}
}

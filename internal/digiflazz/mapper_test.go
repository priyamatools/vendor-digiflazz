package digiflazz

import (
	"testing"

	"github.com/addon-vendor-digiflazz/internal/atrilinks"
)

func TestMapFromUpstreamSuccess(t *testing.T) {
	up := atrilinks.TransactionResponse{
		Status: "success",
		Kode:   "pln50",
		Harga:  51760,
		Saldo:  856841510,
		Bukti:  "0482-3001-6419-9558-7075",
		Info:   "ok success",
	}
	data := MapFromUpstream(up, "16115009", "163000191067", "pln50", 50000)
	if data.Status != "1" || data.RC != "00" {
		t.Fatalf("status/rc: %s %s", data.Status, data.RC)
	}
	if data.SN != up.Bukti || data.Price != "51760" {
		t.Fatalf("sn/price: %s %s", data.SN, data.Price)
	}
	if data.HP != "163000191067" {
		t.Fatalf("hp should stay request hp: %s", data.HP)
	}
}

func TestMapFromUpstreamPending(t *testing.T) {
	up := atrilinks.TransactionResponse{Status: "pending"}
	data := MapFromUpstream(up, "r1", "08123", "xl10", 10000)
	if data.Status != "0" || data.RC != "39" {
		t.Fatalf("got %s %s", data.Status, data.RC)
	}
}

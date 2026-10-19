package service

import (
	"testing"

	"github.com/addon-vendor-digiflazz/internal/atrilinks"
)

func TestShouldApplyCheckToday(t *testing.T) {
	if !shouldApplyCheckToday(atrilinks.TransactionResponse{Status: "success"}) {
		t.Fatal("expected success")
	}
	if !shouldApplyCheckToday(atrilinks.TransactionResponse{Status: "pending"}) {
		t.Fatal("expected pending")
	}
	if shouldApplyCheckToday(atrilinks.TransactionResponse{Status: ""}) {
		t.Fatal("expected empty skip")
	}
}

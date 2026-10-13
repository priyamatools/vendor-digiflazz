package atrilinks

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestExtractCallbackPayloadGETQuery(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/cb?reqid=1&status=success&kode=xl10", nil)
	raw, err := ExtractCallbackPayload(r)
	if err != nil {
		t.Fatal(err)
	}
	tr, err := ParseTransactionResponse(raw)
	if err != nil || tr.ReqID != "1" || tr.Status != "success" {
		t.Fatalf("%v %+v", err, tr)
	}
}

func TestExtractCallbackPayloadPOSTForm(t *testing.T) {
	body := "reqid=2&status=pending&info=process"
	r := httptest.NewRequest(http.MethodPost, "/cb", strings.NewReader(body))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	raw, err := ExtractCallbackPayload(r)
	if err != nil {
		t.Fatal(err)
	}
	tr, err := ParseTransactionResponse(raw)
	if err != nil || tr.ReqID != "2" {
		t.Fatalf("%v %+v", err, tr)
	}
}

func TestExtractCallbackPayloadPOSTJSON(t *testing.T) {
	body := `{"reqid":"3","status":"success","harga":"51760","saldo":100}`
	r := httptest.NewRequest(http.MethodPost, "/cb", strings.NewReader(body))
	raw, err := ExtractCallbackPayload(r)
	if err != nil {
		t.Fatal(err)
	}
	tr, err := ParseTransactionResponse(raw)
	if err != nil || tr.ReqID != "3" || tr.Harga != 51760 {
		t.Fatalf("%v %+v", err, tr)
	}
}

func TestExtractCallbackPayloadGETPayloadParam(t *testing.T) {
	q := url.QueryEscape(`{"reqid":"4","status":"failed","info":"x"}`)
	r := httptest.NewRequest(http.MethodGet, "/cb?payload="+q, nil)
	raw, err := ExtractCallbackPayload(r)
	if err != nil {
		t.Fatal(err)
	}
	tr, err := ParseTransactionResponse(raw)
	if err != nil || tr.ReqID != "4" {
		t.Fatalf("%v %+v", err, tr)
	}
}

package digiflazz

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestWebhookBodyWrappedInData(t *testing.T) {
	body, err := json.Marshal(Envelope{Data: ResponseData{
		RefID: "r1", Status: "1", RC: "00", Price: "3855",
	}})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(body), `"data"`) {
		t.Fatalf("missing data wrapper: %s", body)
	}
	var outer map[string]json.RawMessage
	if err := json.Unmarshal(body, &outer); err != nil {
		t.Fatal(err)
	}
	if _, ok := outer["data"]; !ok {
		t.Fatalf("expected data key: %s", body)
	}
}

package digiflazz

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type WebhookClient struct {
	url    string
	client *http.Client
}

func NewWebhookClient(url string) *WebhookClient {
	return &WebhookClient{
		url: url,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (w *WebhookClient) Send(ctx context.Context, data ResponseData) error {
	body, err := json.Marshal(Envelope{Data: data})
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, w.url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	res, err := w.client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return fmt.Errorf("digiflazz callback status %d", res.StatusCode)
	}
	return nil
}

package digiflazz

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"
)

type WebhookClient struct {
	url    string
	client *http.Client
	log    *slog.Logger
	dev    bool
}

func NewWebhookClient(url string, log *slog.Logger, development bool) *WebhookClient {
	return &WebhookClient{
		url: url,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		log: log,
		dev: development,
	}
}

func (w *WebhookClient) Send(ctx context.Context, data ResponseData) error {
	env := Envelope{Data: data}
	body, err := json.Marshal(env)
	if err != nil {
		return err
	}
	if w.dev {
		w.log.Debug("digiflazz webhook", "body", string(body))
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

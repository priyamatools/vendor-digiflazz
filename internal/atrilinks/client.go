package atrilinks

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Client struct {
	baseURL      string
	userID       string
	pin          string
	pass         string
	urlCallback  string
	httpClient   *http.Client
}

func NewClient(baseURL, userID, pin, pass, urlCallback string) *Client {
	baseURL = strings.TrimRight(baseURL, "/")
	return &Client{
		baseURL:     baseURL,
		userID:      userID,
		pin:         pin,
		pass:        pass,
		urlCallback: urlCallback,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

type OrderParams struct {
	ReqID       string
	KodeProduk  string
	Tujuan      string
	JenisTrx    int
	LimitHarga  int64
}

func (c *Client) Order(ctx context.Context, p OrderParams) ([]byte, TransactionResponse, error) {
	q := url.Values{}
	q.Set("reqid", p.ReqID)
	q.Set("userid", c.userID)
	q.Set("kodeproduk", p.KodeProduk)
	q.Set("tujuan", p.Tujuan)
	q.Set("jenistrx", fmt.Sprintf("%d", p.JenisTrx))
	q.Set("pin", c.pin)
	q.Set("urlcallback", c.urlCallback)
	if c.pass != "" {
		q.Set("pass", c.pass)
	}
	if p.LimitHarga > 0 {
		q.Set("limitharga", fmt.Sprintf("%d", p.LimitHarga))
	}

	u := c.baseURL + "/h2h/trx/query?" + q.Encode()
	return c.doGET(ctx, u)
}

func (c *Client) CheckToday(ctx context.Context, reqID, tujuan string) ([]byte, TransactionResponse, error) {
	q := url.Values{}
	q.Set("reqid", reqID)
	q.Set("userid", c.userID)
	q.Set("tujuan", tujuan)

	u := c.baseURL + "/h2h/trx/checktoday?" + q.Encode()
	return c.doGET(ctx, u)
}

func (c *Client) doGET(ctx context.Context, u string) ([]byte, TransactionResponse, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, TransactionResponse{}, err
	}
	res, err := c.httpClient.Do(req)
	if err != nil {
		return nil, TransactionResponse{}, err
	}
	defer res.Body.Close()

	raw, err := io.ReadAll(io.LimitReader(res.Body, 1<<20))
	if err != nil {
		return nil, TransactionResponse{}, err
	}
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return raw, TransactionResponse{}, fmt.Errorf("atrilinks http %d", res.StatusCode)
	}

	tr, err := ParseTransactionResponse(raw)
	if err != nil {
		return raw, TransactionResponse{}, err
	}
	return raw, tr, nil
}

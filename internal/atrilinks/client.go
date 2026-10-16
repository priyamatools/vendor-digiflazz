package atrilinks

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/addon-vendor-digiflazz/internal/httplog"
)

// Client memanggil API H2H Atrilinks (metode Atrilinks + BLAKE3).
type Client struct {
	baseURL     string
	userID      string
	pin         string
	password    string
	urlCallback string
	httpClient  *http.Client
	log         *slog.Logger
	dev         bool
}

// NewClient membuat klien HTTP Atrilinks dengan kredensial dari environment.
func NewClient(baseURL, userID, password, pin, urlCallback string, log *slog.Logger, development bool) *Client {
	baseURL = strings.TrimRight(baseURL, "/")
	return &Client{
		baseURL:     baseURL,
		userID:      userID,
		pin:         pin,
		password:    password,
		urlCallback: urlCallback,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
		log: log,
		dev: development,
	}
}

// OrderParams parameter transaksi prabayar ke Atrilinks.
type OrderParams struct {
	ReqID      string
	KodeProduk string
	Tujuan     string
	JenisTrx   int
	LimitHarga int64
}

// Order mengirim transaksi fixed (jenistrx=1) ke GET /h2h/trx/atrilinks dengan sign BLAKE3.
func (c *Client) Order(ctx context.Context, p OrderParams) ([]byte, TransactionResponse, error) {
	sign := Sign(p.ReqID, p.KodeProduk, p.Tujuan, c.userID, c.password, c.pin)

	q := url.Values{}
	q.Set("reqid", p.ReqID)
	q.Set("userid", c.userID)
	q.Set("kodeproduk", p.KodeProduk)
	q.Set("tujuan", p.Tujuan)
	q.Set("jenistrx", fmt.Sprintf("%d", p.JenisTrx))
	q.Set("sign", sign)
	q.Set("urlcallback", c.urlCallback)
	if p.LimitHarga > 0 {
		q.Set("limitharga", fmt.Sprintf("%d", p.LimitHarga))
	}

	u := c.baseURL + "/h2h/trx/atrilinks?" + q.Encode()
	if c.dev {
		httplog.AtrilinksOrder(c.log, "/h2h/trx/atrilinks", httplog.RedactAtrilinksURL(u),
			p.ReqID, p.KodeProduk, p.Tujuan, p.JenisTrx, p.LimitHarga)
	}
	return c.doGET(ctx, u, p.ReqID)
}

// CheckToday mengecek status transaksi hari ini (tanpa sign).
func (c *Client) CheckToday(ctx context.Context, reqID, tujuan string) ([]byte, TransactionResponse, error) {
	q := url.Values{}
	q.Set("reqid", reqID)
	q.Set("userid", c.userID)
	q.Set("tujuan", tujuan)

	u := c.baseURL + "/h2h/trx/checktoday?" + q.Encode()
	if c.dev {
		c.log.Debug("atrilinks request",
			"endpoint", "/h2h/trx/checktoday",
			"url", httplog.RedactAtrilinksURL(u),
			"reqid", reqID,
			"tujuan", tujuan,
		)
	}
	return c.doGET(ctx, u, reqID)
}

// doGET menjalankan request GET dan mem-parse response JSON transaksi.
func (c *Client) doGET(ctx context.Context, u string, refID string) ([]byte, TransactionResponse, error) {
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
		if c.dev {
			httplog.AtrilinksResponse(c.log, refID, res.StatusCode, raw, nil, err)
		}
		return nil, TransactionResponse{}, err
	}
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		parseErr := fmt.Errorf("atrilinks http %d", res.StatusCode)
		if c.dev {
			httplog.AtrilinksResponse(c.log, refID, res.StatusCode, raw, nil, parseErr)
		}
		return raw, TransactionResponse{}, parseErr
	}

	tr, err := ParseTransactionResponse(raw)
	if c.dev {
		httplog.AtrilinksResponse(c.log, refID, res.StatusCode, raw, map[string]string{
			"status": tr.Status,
			"jenis":  tr.Jenis,
			"kode":   tr.Kode,
			"tujuan": tr.Tujuan,
			"info":   tr.Info,
			"bukti":  tr.Bukti,
		}, err)
	}
	if err != nil {
		return raw, TransactionResponse{}, err
	}
	return raw, tr, nil
}

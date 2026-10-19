package repository

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("transaction not found")

type Transaction struct {
	ID               int64
	RefID            string
	HP               string
	PulsaCode        string
	Price            int64
	ResponsePrice    string
	DigiflazzUsername string
	DigiflazzStatus  string
	DigiflazzRC      string
	Message          string
	SN               string
	TrID             string
	Balance          string
	AtrilinksStatus  *string
	UpstreamRaw      json.RawMessage
	OrderSubmitted      bool
	CallbackSentAt      *time.Time
	CreatedAt           time.Time
	LastStatusCheckAt   *time.Time
}

type Repository struct {
	pool *pgxpool.Pool
}

// New membuat repository PostgreSQL.
func New(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

// GetByRefID mengambil transaksi berdasarkan ref_id Digiflazz.
func (r *Repository) GetByRefID(ctx context.Context, refID string) (Transaction, error) {
	const q = `
SELECT id, ref_id, hp, pulsa_code, price, response_price, digiflazz_username,
       digiflazz_status, digiflazz_rc, message, sn, tr_id, balance,
       atrilinks_status, upstream_raw, order_submitted, callback_sent_at,
       created_at, last_status_check_at
FROM transactions WHERE ref_id = $1`
	row := r.pool.QueryRow(ctx, q, refID)
	return scanTx(row)
}

// CreatePending menyimpan transaksi baru dengan status pending Digiflazz.
func (r *Repository) CreatePending(ctx context.Context, tx Transaction) error {
	const q = `
INSERT INTO transactions (
  ref_id, hp, pulsa_code, price, digiflazz_username,
  digiflazz_status, digiflazz_rc, message, tr_id, balance, order_submitted
) VALUES ($1,$2,$3,$4,$5,'0','39','Process',$1,'0',false)`
	_, err := r.pool.Exec(ctx, q,
		tx.RefID, tx.HP, tx.PulsaCode, tx.Price, tx.DigiflazzUsername,
	)
	return err
}

type UpstreamUpdate struct {
	RefID           string
	AtrilinksStatus string
	DigiflazzStatus string
	DigiflazzRC     string
	Message         string
	SN              string
	Balance         string
	ResponsePrice   string
	UpstreamRaw     json.RawMessage
	OrderSubmitted  bool
}

// ApplyUpstream memperbarui status hasil upstream dan field response Digiflazz.
func (r *Repository) ApplyUpstream(ctx context.Context, u UpstreamUpdate) error {
	const q = `
UPDATE transactions SET
  atrilinks_status = $2,
  digiflazz_status = $3,
  digiflazz_rc = $4,
  message = $5,
  sn = $6,
  balance = $7,
  response_price = $8,
  upstream_raw = $9,
  order_submitted = $10,
  updated_at = NOW()
WHERE ref_id = $1`
	_, err := r.pool.Exec(ctx, q,
		u.RefID, u.AtrilinksStatus, u.DigiflazzStatus, u.DigiflazzRC,
		u.Message, u.SN, u.Balance, u.ResponsePrice, u.UpstreamRaw, u.OrderSubmitted,
	)
	return err
}

// MarkCallbackSent mencatat waktu webhook Digiflazz berhasil dikirim.
func (r *Repository) MarkCallbackSent(ctx context.Context, refID string) error {
	const q = `UPDATE transactions SET callback_sent_at = NOW(), updated_at = NOW() WHERE ref_id = $1`
	_, err := r.pool.Exec(ctx, q, refID)
	return err
}

// AcquireStatusCheck mengunci slot poll checktoday per ref_id (throttle antar request/worker).
func (r *Repository) AcquireStatusCheck(ctx context.Context, refID string, minInterval time.Duration) (bool, error) {
	sec := int64(minInterval.Seconds())
	if sec < 1 {
		sec = 1
	}
	const q = `
UPDATE transactions SET last_status_check_at = NOW(), updated_at = NOW()
WHERE ref_id = $1
  AND digiflazz_status = '0'
  AND order_submitted = TRUE
  AND (last_status_check_at IS NULL OR last_status_check_at < NOW() - make_interval(secs => $2))
RETURNING ref_id`
	var got string
	err := r.pool.QueryRow(ctx, q, refID, sec).Scan(&got)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

// ListPendingForReconcile mengambil transaksi pending yang layak di-poll checktoday (hari ini di timezone).
func (r *Repository) ListPendingForReconcile(
	ctx context.Context,
	tz string,
	limit int,
	minAge time.Duration,
) ([]Transaction, error) {
	const q = `
SELECT id, ref_id, hp, pulsa_code, price, response_price, digiflazz_username,
       digiflazz_status, digiflazz_rc, message, sn, tr_id, balance,
       atrilinks_status, upstream_raw, order_submitted, callback_sent_at,
       created_at, last_status_check_at
FROM transactions
WHERE digiflazz_status = '0'
  AND order_submitted = TRUE
  AND (created_at AT TIME ZONE $1)::date = (NOW() AT TIME ZONE $1)::date
  AND updated_at < NOW() - make_interval(secs => $2)
ORDER BY COALESCE(last_status_check_at, created_at) ASC
LIMIT $3`
	minSec := int64(minAge.Seconds())
	if minSec < 1 {
		minSec = 1
	}
	rows, err := r.pool.Query(ctx, q, tz, minSec, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Transaction
	for rows.Next() {
		var t Transaction
		if err := rows.Scan(
			&t.ID, &t.RefID, &t.HP, &t.PulsaCode, &t.Price, &t.ResponsePrice, &t.DigiflazzUsername,
			&t.DigiflazzStatus, &t.DigiflazzRC, &t.Message, &t.SN, &t.TrID, &t.Balance,
			&t.AtrilinksStatus, &t.UpstreamRaw, &t.OrderSubmitted, &t.CallbackSentAt,
			&t.CreatedAt, &t.LastStatusCheckAt,
		); err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

// scanTx memetakan satu baris query ke struct Transaction.
func scanTx(row pgx.Row) (Transaction, error) {
	var t Transaction
	err := row.Scan(
		&t.ID, &t.RefID, &t.HP, &t.PulsaCode, &t.Price, &t.ResponsePrice, &t.DigiflazzUsername,
		&t.DigiflazzStatus, &t.DigiflazzRC, &t.Message, &t.SN, &t.TrID, &t.Balance,
		&t.AtrilinksStatus, &t.UpstreamRaw, &t.OrderSubmitted, &t.CallbackSentAt,
		&t.CreatedAt, &t.LastStatusCheckAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return Transaction{}, ErrNotFound
	}
	return t, err
}

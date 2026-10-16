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
	OrderSubmitted   bool
	CallbackSentAt   *time.Time
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
       atrilinks_status, upstream_raw, order_submitted, callback_sent_at
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

// scanTx memetakan satu baris query ke struct Transaction.
func scanTx(row pgx.Row) (Transaction, error) {
	var t Transaction
	err := row.Scan(
		&t.ID, &t.RefID, &t.HP, &t.PulsaCode, &t.Price, &t.ResponsePrice, &t.DigiflazzUsername,
		&t.DigiflazzStatus, &t.DigiflazzRC, &t.Message, &t.SN, &t.TrID, &t.Balance,
		&t.AtrilinksStatus, &t.UpstreamRaw, &t.OrderSubmitted, &t.CallbackSentAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return Transaction{}, ErrNotFound
	}
	return t, err
}

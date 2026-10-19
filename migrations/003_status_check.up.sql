ALTER TABLE transactions
    ADD COLUMN IF NOT EXISTS last_status_check_at TIMESTAMPTZ;

CREATE INDEX IF NOT EXISTS idx_transactions_reconcile
    ON transactions (last_status_check_at)
    WHERE digiflazz_status = '0' AND order_submitted = TRUE;

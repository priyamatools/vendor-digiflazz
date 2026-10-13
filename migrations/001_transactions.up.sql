CREATE TABLE IF NOT EXISTS transactions (
    id                  BIGSERIAL PRIMARY KEY,
    ref_id              TEXT NOT NULL UNIQUE,
    hp                  TEXT NOT NULL,
    pulsa_code          TEXT NOT NULL,
    price               BIGINT NOT NULL,
    digiflazz_username  TEXT NOT NULL,
    digiflazz_status    TEXT NOT NULL DEFAULT '0',
    digiflazz_rc        TEXT NOT NULL DEFAULT '39',
    message             TEXT NOT NULL DEFAULT '',
    sn                  TEXT NOT NULL DEFAULT '',
    tr_id               TEXT NOT NULL DEFAULT '',
    balance             TEXT NOT NULL DEFAULT '0',
    atrilinks_status    TEXT,
    upstream_raw        JSONB,
    order_submitted     BOOLEAN NOT NULL DEFAULT FALSE,
    callback_sent_at    TIMESTAMPTZ,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_transactions_pending
    ON transactions (digiflazz_status)
    WHERE digiflazz_status = '0';

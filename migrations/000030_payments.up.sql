CREATE TYPE payment_verification_status AS ENUM (
    'pending',
    'detected',
    'verified',
    'invalid',
    'unmatched',
    'failed'
);

ALTER TABLE invoice
    ADD COLUMN payment_reference VARCHAR(28) NOT NULL
        DEFAULT ('STL-' || upper(substr(md5(gen_random_uuid()::text), 1, 12))),
    ADD COLUMN recipient_address VARCHAR(56);

CREATE UNIQUE INDEX idx_invoice_payment_reference ON invoice(payment_reference);

UPDATE invoice i
SET recipient_address = w.address
FROM wallets w
WHERE w.user_id = i.created_by_id
  AND w.is_primary = TRUE
  AND w.is_active = TRUE
  AND i.status IN ('sent', 'viewed', 'overdue')
  AND i.recipient_address IS NULL;

CREATE TABLE payments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    invoice_id UUID REFERENCES invoice(id) ON DELETE RESTRICT,
    operation_id VARCHAR(32) NOT NULL UNIQUE,
    transaction_hash VARCHAR(64) NOT NULL,
    sender VARCHAR(56) NOT NULL,
    recipient VARCHAR(56) NOT NULL,
    asset_code VARCHAR(12) NOT NULL,
    asset_issuer VARCHAR(56),
    amount NUMERIC(20, 7) NOT NULL CHECK (amount >= 0),
    payment_reference VARCHAR(28),
    verification_status payment_verification_status NOT NULL DEFAULT 'detected',
    failure_reason TEXT,
    ledger_closed_at TIMESTAMPTZ NOT NULL,
    detected_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    verified_at TIMESTAMPTZ,
    CONSTRAINT payments_verified_at_check
        CHECK ((verification_status = 'verified') = (verified_at IS NOT NULL)),
    CONSTRAINT payments_verified_has_invoice
        CHECK (verification_status <> 'verified' OR invoice_id IS NOT NULL),
    CONSTRAINT payments_rejection_has_reason
        CHECK (verification_status IN ('pending', 'detected', 'verified') OR failure_reason IS NOT NULL)
);

CREATE UNIQUE INDEX uq_payments_one_verified_per_invoice
    ON payments(invoice_id) WHERE verification_status = 'verified';
CREATE INDEX idx_payments_invoice ON payments(invoice_id) WHERE invoice_id IS NOT NULL;
CREATE INDEX idx_payments_transaction_hash ON payments(transaction_hash);
CREATE INDEX idx_payments_recipient_status ON payments(recipient, verification_status);
CREATE INDEX idx_payments_unfinished ON payments(detected_at)
    WHERE verification_status IN ('pending', 'detected');

CREATE TABLE payment_cursors (
    network VARCHAR(8) NOT NULL,
    account VARCHAR(56) NOT NULL,
    cursor TEXT NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (network, account)
);

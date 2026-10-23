ALTER TABLE payments
    ADD COLUMN failure_code TEXT,
    ADD COLUMN checks JSONB,
    ADD COLUMN expected JSONB;

UPDATE payments SET failure_code = CASE
    WHEN failure_reason LIKE 'insufficient amount%' THEN 'insufficient_amount'
    WHEN failure_reason LIKE 'amount mismatch%' THEN 'amount_mismatch'
    WHEN failure_reason LIKE 'expected asset%' THEN 'wrong_asset'
    WHEN failure_reason LIKE '%recipient is not%' THEN 'wrong_recipient'
    WHEN failure_reason LIKE '%no text memo%' THEN 'missing_reference'
    WHEN failure_reason LIKE 'no invoice matches%' THEN 'unknown_reference'
    WHEN failure_reason LIKE 'transaction failed%' THEN 'transaction_failed'
    WHEN failure_reason LIKE 'invoice is already paid%' THEN 'duplicate'
    WHEN failure_reason LIKE '%payment window has expired%' THEN 'expired'
    WHEN failure_reason LIKE 'invoice is % and cannot accept payment' THEN 'invoice_not_payable'
    WHEN failure_reason LIKE 'invoice total is zero%' THEN 'invalid_invoice_total'
END
WHERE failure_reason IS NOT NULL;

-- Duplicates were stored as unmatched when the invoice had no recipient.
UPDATE invoice i SET recipient_address = w.address
FROM wallets w
WHERE i.recipient_address IS NULL
  AND w.user_id = i.created_by_id AND w.chain = 'stellar' AND w.is_primary AND w.is_active;

UPDATE payments p SET verification_status = 'invalid', failure_code = 'duplicate',
       failure_reason = 'invoice is already paid', invoice_id = i.id
FROM invoice i
WHERE p.verification_status = 'unmatched' AND p.payment_reference = i.payment_reference
  AND i.status = 'paid' AND p.recipient = i.recipient_address;

-- Ledger status is the on-chain result; verification lives on the payment record. One type for incoming funds.
UPDATE transactions t SET status = 'confirmed', transaction_type = 'payment'
FROM payments p
WHERE p.transaction_hash = t.transaction_hash
  AND p.verification_status IN ('invalid', 'unmatched')
  AND t.transaction_type IN ('funding', 'payment');

UPDATE transactions t SET transaction_type = 'payment'
FROM payments p
WHERE p.transaction_hash = t.transaction_hash AND p.verification_status = 'failed' AND t.transaction_type = 'funding';

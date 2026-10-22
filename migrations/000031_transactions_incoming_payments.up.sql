ALTER TABLE transactions ALTER COLUMN invoice_id DROP NOT NULL;

INSERT INTO transactions (invoice_id, wallet_id, transaction_hash, user_id, amount, currency, status,
                          token_type, transaction_type, confirmed_at)
SELECT p.invoice_id, w.id, p.transaction_hash, w.user_id, p.amount,
       lower(p.asset_code)::currency_type,
       CASE p.verification_status WHEN 'unmatched' THEN 'confirmed' ELSE 'failed' END::transaction_status,
       lower(p.asset_code)::currency_type,
       CASE p.verification_status WHEN 'unmatched' THEN 'funding' ELSE 'payment' END::transaction_type,
       p.ledger_closed_at
FROM payments p
JOIN wallets w ON w.address = p.recipient AND w.chain = 'stellar'
WHERE p.verification_status IN ('unmatched', 'invalid', 'failed')
  AND lower(p.asset_code) IN ('usdc', 'xlm')
ON CONFLICT (transaction_hash) DO NOTHING;

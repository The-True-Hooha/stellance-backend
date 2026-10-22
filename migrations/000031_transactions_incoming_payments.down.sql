DELETE FROM transactions WHERE invoice_id IS NULL;
ALTER TABLE transactions ALTER COLUMN invoice_id SET NOT NULL;

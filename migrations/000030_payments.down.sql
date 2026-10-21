DROP TABLE IF EXISTS payment_cursors;
DROP TABLE IF EXISTS payments;
DROP INDEX IF EXISTS idx_invoice_payment_reference;
ALTER TABLE invoice DROP COLUMN IF EXISTS recipient_address, DROP COLUMN IF EXISTS payment_reference;
DROP TYPE IF EXISTS payment_verification_status;

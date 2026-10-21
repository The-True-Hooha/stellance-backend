ALTER TYPE invoice_status ADD VALUE IF NOT EXISTS 'payment_detected';
ALTER TYPE invoice_status ADD VALUE IF NOT EXISTS 'invalid_payment';
ALTER TYPE invoice_status ADD VALUE IF NOT EXISTS 'expired';

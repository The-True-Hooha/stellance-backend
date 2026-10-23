package payments

import (
	"context"
	"time"

	"github.com/stellar/go/amount"
)

func (s *PaymentService) paymentRequest(ctx context.Context, where string, key any) (*PaymentRequest, error) {
	_, stage, err := s.chain(ctx)
	if err != nil {
		return nil, err
	}

	var (
		r         PaymentRequest
		expiresAt *time.Time
	)
	err = s.postgres.QueryRow(ctx, `
		SELECT i.id, i.invoice_number, i.status::text, i.total::text, i.currency::text,
		       COALESCE(i.recipient_address, (SELECT w.address FROM wallets w
		           WHERE w.user_id = i.created_by_id AND w.chain = 'stellar' AND w.is_primary AND w.is_active LIMIT 1), ''),
		       i.payment_reference, ((i.due_date + $2::int + 1)::timestamp AT TIME ZONE 'UTC'), i.paid_at,
		       lp.transaction_hash, lp.verification_status, lp.detected_at, lp.verified_at
		FROM invoice i
		LEFT JOIN LATERAL (
			SELECT transaction_hash, verification_status::text, detected_at, verified_at
			FROM payments p WHERE p.invoice_id = i.id
			ORDER BY (p.verification_status = 'verified') DESC, p.detected_at DESC LIMIT 1
		) lp ON TRUE
		WHERE `+where, key, graceDays()).Scan(
		&r.InvoiceID, &r.InvoiceNumber, &r.InvoiceStatus, &r.Amount, &r.currency,
		&r.Recipient, &r.PaymentReference, &expiresAt, &r.PaidAt,
		&r.TransactionHash, &r.VerificationStatus, &r.PaymentDetectedAt, &r.VerifiedAt)
	if err != nil {
		return nil, err
	}

	r.Amount = normalizeAmount(r.Amount)
	a := expectedAsset(stage, r.currency)
	r.AssetCode = a.Code
	if a.Issuer != "" {
		r.AssetIssuer = &a.Issuer
	}
	r.Network = stage
	r.ExpiresAt = expiresAt
	r.PaymentStatus = PaymentStatusFor(r.InvoiceStatus)
	if r.TransactionHash != nil && r.VerificationStatus != nil && *r.VerificationStatus == string(StatusVerified) {
		u := explorerTxURL(stage, *r.TransactionHash)
		r.ExplorerURL = &u
	} else {
		r.TransactionHash = nil
	}
	return &r, nil
}

const paymentColumns = `
	SELECT p.id, p.invoice_id, i.invoice_number, p.transaction_hash, p.sender, p.recipient, p.asset_code, p.asset_issuer,
	       p.amount::text, p.payment_reference, p.verification_status::text, p.failure_code, p.failure_reason,
	       p.checks, p.expected, p.ledger_closed_at, p.detected_at, p.verified_at, `

const paymentFrom = `
	FROM payments p LEFT JOIN invoice i ON i.id = p.invoice_id `

const (
	paymentSelect      = paymentColumns + `NULL::text` + paymentFrom
	adminPaymentSelect = paymentColumns + `(SELECT u.email FROM wallets w JOIN users u ON u.id = w.user_id
	                       WHERE w.address = p.recipient AND w.chain = 'stellar' LIMIT 1)` + paymentFrom
)

type rowScanner interface {
	Scan(dest ...any) error
}

func scanPayment(row rowScanner, stage string) (PaymentRecord, error) {
	var (
		r   PaymentRecord
		amt string
	)
	if err := row.Scan(&r.ID, &r.InvoiceID, &r.InvoiceNumber, &r.TransactionHash, &r.Sender, &r.Recipient,
		&r.AssetCode, &r.AssetIssuer, &amt, &r.PaymentReference, &r.VerificationStatus, &r.FailureCode,
		&r.FailureReason, &r.Checks, &r.Expected, &r.LedgerClosedAt, &r.DetectedAt, &r.VerifiedAt,
		&r.MerchantEmail); err != nil {
		return r, err
	}
	r.Amount = normalizeAmount(amt)
	r.ExplorerURL = explorerTxURL(stage, r.TransactionHash)
	return r, nil
}

// normalizeAmount renders a decimal string with the 7 fractional digits Stellar uses.
func normalizeAmount(v string) string {
	st, err := amount.ParseInt64(v)
	if err != nil {
		return v
	}
	return amount.StringFromInt64(st)
}

func (s *PaymentService) queryPayments(ctx context.Context, stage, where string, args ...any) ([]PaymentRecord, error) {
	return s.queryPaymentsFrom(ctx, paymentSelect, stage, where, args...)
}

func (s *PaymentService) queryPaymentsFrom(ctx context.Context, sel, stage, where string, args ...any) ([]PaymentRecord, error) {
	rows, err := s.postgres.Query(ctx, sel+where, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []PaymentRecord{}
	for rows.Next() {
		r, err := scanPayment(rows, stage)
		if err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

func (s *PaymentService) paymentsByHash(ctx context.Context, hash, stage string) ([]PaymentRecord, error) {
	return s.queryPayments(ctx, stage, `WHERE p.transaction_hash = $1 ORDER BY p.detected_at`, hash)
}

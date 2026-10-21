package payments

import (
	"context"
	"strings"
	"time"
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
		       COALESCE(i.recipient_address, ''), i.payment_reference, ((i.due_date + $2::int + 1)::timestamp AT TIME ZONE 'UTC'), i.paid_at,
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

	a := expectedAsset(stage, r.currency)
	r.Asset = a.Code
	if a.Issuer != "" {
		r.AssetIssuer = &a.Issuer
	}
	r.Network = stage
	r.ExpiresAt = expiresAt
	r.PaymentStatus = paymentStatusFor(r.InvoiceStatus)
	if r.TransactionHash != nil && r.VerificationStatus != nil && *r.VerificationStatus == string(StatusVerified) {
		u := explorerTxURL(stage, *r.TransactionHash)
		r.ExplorerURL = &u
	} else {
		r.TransactionHash = nil
	}
	return &r, nil
}

const paymentSelect = `
	SELECT p.id, p.invoice_id, p.transaction_hash, p.sender, p.recipient, p.asset_code, p.asset_issuer,
	       p.amount::text, p.payment_reference, p.verification_status::text, p.failure_reason,
	       p.ledger_closed_at, p.detected_at, p.verified_at
	FROM payments p `

type rowScanner interface {
	Scan(dest ...any) error
}

func scanPayment(row rowScanner, stage string) (PaymentRecord, error) {
	var (
		r      PaymentRecord
		code   string
		issuer *string
		amt    string
	)
	if err := row.Scan(&r.ID, &r.InvoiceID, &r.TransactionHash, &r.Sender, &r.Recipient, &code, &issuer,
		&amt, &r.PaymentReference, &r.VerificationStatus, &r.FailureReason,
		&r.LedgerClosedAt, &r.DetectedAt, &r.VerifiedAt); err != nil {
		return r, err
	}
	a := asset{Code: code}
	if issuer != nil {
		a.Issuer = *issuer
	}
	r.Asset = a.String()
	r.Amount = trimAmount(amt)
	r.ExplorerURL = explorerTxURL(stage, r.TransactionHash)
	return r, nil
}

func trimAmount(v string) string {
	if strings.Contains(v, ".") {
		v = strings.TrimRight(strings.TrimRight(v, "0"), ".")
	}
	return v
}

func (s *PaymentService) queryPayments(ctx context.Context, stage, where string, args ...any) ([]PaymentRecord, error) {
	rows, err := s.postgres.Query(ctx, paymentSelect+where, args...)
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

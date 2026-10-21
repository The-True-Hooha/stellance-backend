package payments

import (
	"context"
	"errors"
	"net/http"
	"slices"

	"github.com/The-True-Hooha/stellance-backend/pkg/utils"
	"github.com/jackc/pgx/v5"
)

func (s *PaymentService) internalError(msg string, err error, args ...any) *utils.ApiResponse {
	s.log.Error(msg, append(args, "error", err)...)
	return &utils.ApiResponse{StatusCode: http.StatusInternalServerError, Message: "failed to process request"}
}

// GetInvoicePayments returns the payment request and every payment recorded against it to the invoice owner.
func (s *PaymentService) GetInvoicePayments(ctx context.Context, userID, invoiceID string) *utils.ApiResponse {
	var owner string
	err := s.postgres.QueryRow(ctx, `SELECT created_by_id FROM invoice WHERE id = $1`, invoiceID).Scan(&owner)
	if errors.Is(err, pgx.ErrNoRows) {
		return &utils.ApiResponse{StatusCode: http.StatusNotFound, Message: "invoice not found"}
	}
	if err != nil {
		return s.internalError("load invoice owner", err, "invoice_id", invoiceID)
	}
	if owner != userID {
		return &utils.ApiResponse{StatusCode: http.StatusForbidden, Message: "access denied"}
	}

	req, err := s.EnsurePaymentRequest(ctx, invoiceID)
	if errors.Is(err, ErrNoRecipientWallet) {
		return &utils.ApiResponse{StatusCode: http.StatusConflict, Message: err.Error()}
	}
	if errors.Is(err, ErrInvoiceNotPayable) {
		req, err = s.paymentRequest(ctx, "i.id = $1", invoiceID)
	}
	if err != nil {
		return s.internalError("load payment request", err, "invoice_id", invoiceID)
	}

	req.Payments, err = s.queryPayments(ctx, req.Network, `WHERE p.invoice_id = $1 ORDER BY p.detected_at DESC`, invoiceID)
	if err != nil {
		return s.internalError("load invoice payments", err, "invoice_id", invoiceID)
	}
	return &utils.ApiResponse{StatusCode: http.StatusOK, Message: "payment details", Data: req}
}

// GetPublicRequest returns what a payer needs to pay an invoice identified by its short url.
func (s *PaymentService) GetPublicRequest(ctx context.Context, invoiceURL string) *utils.ApiResponse {
	var invoiceID, status string
	err := s.postgres.QueryRow(ctx, `SELECT id, status::text FROM invoice WHERE invoice_url = $1`, invoiceURL).Scan(&invoiceID, &status)
	if errors.Is(err, pgx.ErrNoRows) || status == "draft" {
		return &utils.ApiResponse{StatusCode: http.StatusNotFound, Message: "invoice not found"}
	}
	if err != nil {
		return s.internalError("load invoice by url", err)
	}

	req, err := s.EnsurePaymentRequest(ctx, invoiceID)
	if errors.Is(err, ErrNoRecipientWallet) {
		return &utils.ApiResponse{StatusCode: http.StatusConflict, Message: "this invoice cannot accept payment yet"}
	}
	if errors.Is(err, ErrInvoiceNotPayable) {
		req, err = s.paymentRequest(ctx, "i.id = $1", invoiceID)
	}
	if err != nil {
		return s.internalError("load payment request", err, "invoice_id", invoiceID)
	}
	return &utils.ApiResponse{StatusCode: http.StatusOK, Message: "payment request", Data: req}
}

// ListPayments returns payments received on the user's wallets, including unmatched and invalid ones.
func (s *PaymentService) ListPayments(ctx context.Context, userID, status string, limit, offset int) *utils.ApiResponse {
	_, stage, err := s.chain(ctx)
	if err != nil {
		return s.internalError("resolve network", err)
	}

	records, err := s.queryPayments(ctx, stage, `
		WHERE p.recipient IN (SELECT address FROM wallets WHERE user_id = $1)
		  AND ($2 = '' OR p.verification_status::text = $2)
		ORDER BY p.detected_at DESC LIMIT $3 OFFSET $4`, userID, status, limit, offset)
	if err != nil {
		return s.internalError("list payments", err, "user_id", userID)
	}
	return &utils.ApiResponse{StatusCode: http.StatusOK, Message: "payments", Data: records}
}

// VerifyTransactionForUser re-checks a transaction on chain. It reports the payments found and never
// trusts the caller about their outcome.
func (s *PaymentService) VerifyTransactionForUser(ctx context.Context, userID, hash string) *utils.ApiResponse {
	records, err := s.VerifyTransaction(ctx, hash)
	if err != nil {
		s.log.Error("verify transaction", "transaction_hash", hash, "error", err)
		return &utils.ApiResponse{StatusCode: http.StatusBadGateway, Message: "could not verify the transaction on Stellar, try again shortly"}
	}

	rows, err := s.postgres.Query(ctx, `SELECT address FROM wallets WHERE user_id = $1`, userID)
	if err != nil {
		return s.internalError("load user wallets", err, "user_id", userID)
	}
	mine, err := pgx.CollectRows(rows, pgx.RowTo[string])
	if err != nil {
		return s.internalError("load user wallets", err, "user_id", userID)
	}
	visible := []PaymentRecord{}
	for _, r := range records {
		if slices.Contains(mine, r.Recipient) {
			visible = append(visible, r)
		}
	}
	return &utils.ApiResponse{StatusCode: http.StatusOK, Message: "transaction checked", Data: visible}
}

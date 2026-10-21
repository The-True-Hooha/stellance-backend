package payments

import "time"

type PaymentRecord struct {
	ID                 string     `json:"id"`
	InvoiceID          *string    `json:"invoice_id,omitempty"`
	TransactionHash    string     `json:"transaction_hash"`
	ExplorerURL        string     `json:"explorer_url"`
	Sender             string     `json:"sender"`
	Recipient          string     `json:"recipient"`
	Asset              string     `json:"asset"`
	Amount             string     `json:"amount"`
	PaymentReference   *string    `json:"payment_reference,omitempty"`
	VerificationStatus string     `json:"verification_status"`
	FailureReason      *string    `json:"failure_reason,omitempty"`
	LedgerClosedAt     time.Time  `json:"ledger_closed_at"`
	DetectedAt         time.Time  `json:"detected_at"`
	VerifiedAt         *time.Time `json:"verified_at,omitempty"`
}

// PaymentRequest is what a payer needs to pay an invoice and what the owner needs to track it.
type PaymentRequest struct {
	InvoiceID          string          `json:"invoice_id"`
	InvoiceNumber      string          `json:"invoice_number"`
	InvoiceStatus      string          `json:"invoice_status"`
	PaymentStatus      string          `json:"payment_status"`
	Amount             string          `json:"amount"`
	Asset              string          `json:"asset"`
	AssetIssuer        *string         `json:"asset_issuer,omitempty"`
	Recipient          string          `json:"recipient"`
	PaymentReference   string          `json:"payment_reference"`
	Network            string          `json:"network"`
	ExpiresAt          *time.Time      `json:"expires_at,omitempty"`
	PaymentDetectedAt  *time.Time      `json:"payment_detected_at,omitempty"`
	VerificationStatus *string         `json:"verification_status,omitempty"`
	VerifiedAt         *time.Time      `json:"verified_at,omitempty"`
	PaidAt             *time.Time      `json:"paid_at,omitempty"`
	TransactionHash    *string         `json:"transaction_hash,omitempty"`
	ExplorerURL        *string         `json:"explorer_url,omitempty"`
	Payments           []PaymentRecord `json:"payments,omitempty"`

	currency string
}

type VerifyTransactionDTO struct {
	TransactionHash string `json:"transaction_hash" validate:"required,len=64,hexadecimal"`
}

func paymentStatusFor(invoiceStatus string) string {
	switch invoiceStatus {
	case "paid", "payment_detected", "invalid_payment", "expired", "cancelled", "draft", "refunded":
		return invoiceStatus
	default:
		return "awaiting_payment"
	}
}

package payments

import "time"

type PaymentRecord struct {
	ID                 string     `json:"id"`
	InvoiceID          *string    `json:"invoice_id,omitempty"`
	InvoiceNumber      *string    `json:"invoice_number"`
	TransactionHash    string     `json:"transaction_hash"`
	ExplorerURL        string     `json:"explorer_url"`
	Sender             string     `json:"sender"`
	Recipient          string     `json:"recipient"`
	AssetCode          string     `json:"asset_code"`
	AssetIssuer        *string    `json:"asset_issuer"`
	Amount             string     `json:"amount"`
	PaymentReference   *string    `json:"payment_reference"`
	VerificationStatus string     `json:"verification_status"`
	FailureCode        *string    `json:"failure_code"`
	FailureReason      *string    `json:"failure_reason"`
	Checks             *Checks    `json:"checks"`
	Expected           *Expected  `json:"expected"`
	LedgerClosedAt     time.Time  `json:"ledger_closed_at"`
	DetectedAt         time.Time  `json:"detected_at"`
	VerifiedAt         *time.Time `json:"verified_at"`
	MerchantEmail      *string    `json:"merchant_email,omitempty"`
}

// PaymentRequest is what a payer needs to pay an invoice and what the owner needs to track it.
type PaymentRequest struct {
	InvoiceID          string          `json:"invoice_id"`
	InvoiceNumber      string          `json:"invoice_number"`
	InvoiceStatus      string          `json:"invoice_status"`
	PaymentStatus      string          `json:"payment_status"`
	Amount             string          `json:"amount"`
	AssetCode          string          `json:"asset_code"`
	AssetIssuer        *string         `json:"asset_issuer"`
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

func PaymentStatusFor(invoiceStatus string) string {
	switch invoiceStatus {
	case "paid", "payment_detected", "invalid_payment", "expired", "cancelled", "draft", "refunded":
		return invoiceStatus
	default:
		return "awaiting_payment"
	}
}

type PageMeta struct {
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}

type PaymentList struct {
	Payments []PaymentRecord `json:"payments"`
	Meta     PageMeta        `json:"meta"`
}

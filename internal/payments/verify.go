package payments

import (
	"fmt"
	"time"
)

type VerificationStatus string

const (
	StatusPending   VerificationStatus = "pending"
	StatusDetected  VerificationStatus = "detected"
	StatusVerified  VerificationStatus = "verified"
	StatusInvalid   VerificationStatus = "invalid"
	StatusUnmatched VerificationStatus = "unmatched"
	StatusFailed    VerificationStatus = "failed"
)

const (
	invoiceDetected = "payment_detected"
	invoiceInvalid  = "invalid_payment"
	invoicePaid     = "paid"
)

// payableInvoiceStatuses are the invoice states in which an incoming payment may complete the invoice.
var payableInvoiceStatuses = map[string]bool{
	"sent":          true,
	"viewed":        true,
	"overdue":       true,
	"pending":       true,
	invoiceDetected: true,
	invoiceInvalid:  true,
}

type asset struct {
	Code   string
	Issuer string
}

func (a asset) String() string {
	if a.Issuer == "" {
		return a.Code
	}
	return a.Code + ":" + a.Issuer
}

// chainPayment is one incoming payment operation, amounts in stroops (1e-7).
type chainPayment struct {
	OperationID string
	TxHash      string
	From        string
	To          string
	Asset       asset
	Stroops     int64
	Memo        string
	Successful  bool
	ClosedAt    time.Time
}

type invoiceView struct {
	ID        string
	Status    string
	Recipient string
	Reference string
	Currency  string
	Stroops   int64
	// ExpiresAt is exclusive: a payment closed at or after it is late. Judged on ledger close time, not processing time.
	ExpiresAt time.Time
}

type verdict struct {
	Status VerificationStatus
	Reason string
}

// evaluate is the single rule set deciding what an on-chain payment means for an invoice.
// inv is nil when no invoice carries the payment's memo.
func evaluate(p chainPayment, inv *invoiceView, expected asset) verdict {
	switch {
	case !p.Successful:
		return verdict{StatusFailed, "transaction failed on Stellar"}
	case p.Memo == "":
		return verdict{StatusUnmatched, "payment has no text memo reference"}
	case inv == nil:
		return verdict{StatusUnmatched, "no invoice matches the payment reference"}
	case inv.Recipient == "" || p.To != inv.Recipient:
		return verdict{StatusUnmatched, "payment recipient is not the invoice recipient"}
	case inv.Status == invoicePaid:
		return verdict{StatusInvalid, "invoice is already paid"}
	case !payableInvoiceStatuses[inv.Status]:
		return verdict{StatusInvalid, fmt.Sprintf("invoice is %s and cannot accept payment", inv.Status)}
	case !inv.ExpiresAt.IsZero() && !p.ClosedAt.Before(inv.ExpiresAt):
		return verdict{StatusInvalid, "invoice payment window has expired"}
	case inv.Stroops <= 0:
		return verdict{StatusInvalid, "invoice total is zero"}
	case p.Asset != expected:
		return verdict{StatusInvalid, fmt.Sprintf("expected asset %s, received %s", expected, p.Asset)}
	case p.Stroops < inv.Stroops:
		return verdict{StatusInvalid, fmt.Sprintf("insufficient amount: expected %d stroops, received %d", inv.Stroops, p.Stroops)}
	case p.Stroops > inv.Stroops:
		return verdict{StatusInvalid, fmt.Sprintf("amount mismatch: expected %d stroops, received %d", inv.Stroops, p.Stroops)}
	}
	return verdict{StatusVerified, ""}
}

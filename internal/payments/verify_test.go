package payments

import (
	"testing"
	"time"
)

var (
	usdc = asset{Code: "USDC", Issuer: "GISSUER"}
	now  = time.Date(2026, 10, 20, 12, 0, 0, 0, time.UTC)
)

func goodPayment() chainPayment {
	return chainPayment{
		OperationID: "1", TxHash: "h", From: "GPAYER", To: "GVENDOR", Asset: usdc,
		Stroops: 100_0000000, Memo: "STL-ABC", Successful: true, ClosedAt: now,
	}
}

func goodInvoice() *invoiceView {
	return &invoiceView{
		ID: "inv", Status: "sent", Recipient: "GVENDOR", Reference: "STL-ABC",
		Currency: "usdc", Stroops: 100_0000000, ExpiresAt: now.Add(24 * time.Hour),
	}
}

func TestEvaluate(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*chainPayment, *invoiceView) *invoiceView
		want   VerificationStatus
	}{
		{"correct payment", func(p *chainPayment, i *invoiceView) *invoiceView { return i }, StatusVerified},
		{"overpayment", func(p *chainPayment, i *invoiceView) *invoiceView { p.Stroops++; return i }, StatusInvalid},
		{"insufficient", func(p *chainPayment, i *invoiceView) *invoiceView { p.Stroops--; return i }, StatusInvalid},
		{"wrong asset code", func(p *chainPayment, i *invoiceView) *invoiceView { p.Asset.Code = "EURC"; return i }, StatusInvalid},
		{"wrong asset issuer", func(p *chainPayment, i *invoiceView) *invoiceView { p.Asset.Issuer = "GFAKE"; return i }, StatusInvalid},
		{"native instead of usdc", func(p *chainPayment, i *invoiceView) *invoiceView { p.Asset = asset{Code: "XLM"}; return i }, StatusInvalid},
		{"wrong recipient", func(p *chainPayment, i *invoiceView) *invoiceView { p.To = "GOTHER"; return i }, StatusUnmatched},
		{"invoice has no recipient", func(p *chainPayment, i *invoiceView) *invoiceView { i.Recipient = ""; return i }, StatusUnmatched},
		{"unmatched: unknown reference", func(p *chainPayment, i *invoiceView) *invoiceView { return nil }, StatusUnmatched},
		{"unmatched: no memo", func(p *chainPayment, i *invoiceView) *invoiceView { p.Memo = ""; return i }, StatusUnmatched},
		{"failed transaction", func(p *chainPayment, i *invoiceView) *invoiceView { p.Successful = false; return i }, StatusFailed},
		{"failed transaction without invoice", func(p *chainPayment, i *invoiceView) *invoiceView { p.Successful = false; return nil }, StatusFailed},
		{"already paid", func(p *chainPayment, i *invoiceView) *invoiceView { i.Status = "paid"; return i }, StatusInvalid},
		{"cancelled", func(p *chainPayment, i *invoiceView) *invoiceView { i.Status = "cancelled"; return i }, StatusInvalid},
		{"draft", func(p *chainPayment, i *invoiceView) *invoiceView { i.Status = "draft"; return i }, StatusInvalid},
		{"expired status", func(p *chainPayment, i *invoiceView) *invoiceView { i.Status = "expired"; return i }, StatusInvalid},
		{"closed at window end", func(p *chainPayment, i *invoiceView) *invoiceView { i.ExpiresAt = now; return i }, StatusInvalid},
		{"closed just before window end", func(p *chainPayment, i *invoiceView) *invoiceView { i.ExpiresAt = now.Add(time.Second); return i }, StatusVerified},
		{"zero total invoice", func(p *chainPayment, i *invoiceView) *invoiceView { i.Stroops = 0; return i }, StatusInvalid},
		{"retry after invalid payment", func(p *chainPayment, i *invoiceView) *invoiceView { i.Status = "invalid_payment"; return i }, StatusVerified},
		{"in-flight sibling payment", func(p *chainPayment, i *invoiceView) *invoiceView { i.Status = "payment_detected"; return i }, StatusVerified},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			p, inv := goodPayment(), goodInvoice()
			inv = tc.mutate(&p, inv)
			got := evaluate(p, inv, usdc)
			if got.Status != tc.want {
				t.Fatalf("status = %s (%s), want %s", got.Status, got.Reason, tc.want)
			}
			if got.Status != StatusVerified && got.Reason == "" {
				t.Fatal("rejected payment has no reason")
			}
		})
	}
}

package payments

import (
	"testing"

	"github.com/stellar/go/protocols/horizon/operations"
	"github.com/stellar/go/xdr"
)

func parseOp(t *testing.T, typeI xdr.OperationType, body string) (chainPayment, bool) {
	t.Helper()
	op, err := operations.UnmarshalOperation(int32(typeI), []byte(body))
	if err != nil {
		t.Fatal(err)
	}
	return toChainPayment(op)
}

func TestToChainPayment(t *testing.T) {
	const tx = `"transaction_hash":"abc","transaction_successful":true,"created_at":"2026-10-20T12:00:00Z",
		"transaction":{"hash":"abc","memo_type":"text","memo":"STL-ABC123"}`

	p, ok := parseOp(t, xdr.OperationTypePayment, `{"id":"77","type":"payment","type_i":1,`+tx+`,
		"from":"GFROM","to":"GTO","amount":"100.0000000","asset_type":"credit_alphanum4","asset_code":"USDC","asset_issuer":"GISS"}`)
	if !ok || p.OperationID != "77" || p.To != "GTO" || p.Stroops != 100_0000000 || p.Memo != "STL-ABC123" ||
		p.Asset != (asset{Code: "USDC", Issuer: "GISS"}) || !p.Successful || p.TxHash != "abc" {
		t.Fatalf("payment parsed wrong: %+v ok=%v", p, ok)
	}

	p, ok = parseOp(t, xdr.OperationTypePayment, `{"id":"78","type":"payment","type_i":1,`+tx+`,
		"from":"GFROM","to":"GTO","amount":"1.5000000","asset_type":"native"}`)
	if !ok || p.Asset != (asset{Code: "XLM"}) || p.Stroops != 1_5000000 {
		t.Fatalf("native payment parsed wrong: %+v", p)
	}

	p, ok = parseOp(t, xdr.OperationTypePathPaymentStrictReceive, `{"id":"79","type":"path_payment_strict_receive","type_i":2,`+tx+`,
		"from":"GFROM","to":"GTO","amount":"10.0000000","asset_type":"credit_alphanum4","asset_code":"USDC","asset_issuer":"GISS",
		"source_amount":"80.0000000","source_asset_type":"native"}`)
	if !ok || p.Stroops != 10_0000000 || p.Asset.Code != "USDC" {
		t.Fatalf("path payment must report the received amount and asset: %+v", p)
	}

	if _, ok := parseOp(t, xdr.OperationTypeCreateAccount, `{"id":"80","type":"create_account","type_i":0,`+tx+`,
		"funder":"GF","account":"GA","starting_balance":"10.0000000"}`); ok {
		t.Fatal("create_account must not count as a payment")
	}

	p, _ = parseOp(t, xdr.OperationTypePayment, `{"id":"81","type":"payment","type_i":1,"transaction_hash":"abc","transaction_successful":false,
		"created_at":"2026-10-20T12:00:00Z","transaction":{"hash":"abc","memo_type":"id","memo":"123"},
		"from":"GFROM","to":"GTO","amount":"1.0000000","asset_type":"native"}`)
	if p.Successful || p.Memo != "" {
		t.Fatalf("failed tx / non-text memo parsed wrong: %+v", p)
	}

	p, _ = parseOp(t, xdr.OperationTypePayment, `{"id":"82","type":"payment","type_i":1,"transaction_hash":"abc","transaction_successful":true,
		"created_at":"2026-10-20T12:00:00Z","transaction":{"hash":"abc","memo_type":"text","memo":"a\u0000b"},
		"from":"GFROM","to":"GTO","amount":"1.0000000","asset_type":"native"}`)
	if p.Memo != "" {
		t.Fatalf("NUL memo must be dropped, got %q", p.Memo)
	}
}

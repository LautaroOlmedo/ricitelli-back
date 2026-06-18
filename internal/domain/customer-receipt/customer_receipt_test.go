package customer_receipt_test

import (
	"testing"

	customer_receipt "ricitelli-back/internal/domain/customer-receipt"
)

func validParams() customer_receipt.NewCustomerReceiptParams {
	return customer_receipt.NewCustomerReceiptParams{
		CustomerID: "customer-1", ReceiptNumber: 7, ReceiptDate: "2026-06-09", Currency: "ARS",
		ExchangeRate: "1.000000", Amount: "100.0000", PaymentMethod: "TRANSFER",
		IdempotencyKey: "receipt-key",
		Allocations:    []customer_receipt.Allocation{{SalesInvoiceID: "invoice-1", AllocatedAmount: "60.0000"}},
	}
}

func TestCustomerReceiptRejectsOverAllocation(t *testing.T) {
	params := validParams()
	params.Allocations[0].AllocatedAmount = "100.0001"
	if _, err := customer_receipt.NewCustomerReceipt(params); err == nil {
		t.Fatal("NewCustomerReceipt() expected over-allocation error")
	}
}

func TestCustomerReceiptPostIsIdempotentAndImmutable(t *testing.T) {
	receipt, err := customer_receipt.NewCustomerReceipt(validParams())
	if err != nil {
		t.Fatal(err)
	}
	if err := receipt.Post(); err != nil {
		t.Fatal(err)
	}
	if err := receipt.Post(); err != nil {
		t.Fatalf("second Post() error = %v", err)
	}
	if receipt.IsEditable() {
		t.Fatal("posted receipt must not be editable")
	}
}

package purchasing_administration

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDecimalArithmeticIsExact(t *testing.T) {
	sum, err := AddDecimals("0.1", "0.2", "10.0000")
	require.NoError(t, err)
	require.Equal(t, "10.3", sum)
	diff, err := SubtractDecimals("100.0001", "0.0001")
	require.NoError(t, err)
	require.Equal(t, "100", diff)
}

func TestSupplierInvoiceIssueAndVoidLifecycle(t *testing.T) {
	invoice, err := NewSupplierInvoice(NewSupplierInvoiceParams{
		SupplierID: "supplier", DocumentType: SupplierInvoiceTypeInvoice, PointOfSale: 1,
		DocumentNumber: 10, IssueDate: "2026-06-09", Currency: CurrencyARS, ExchangeRate: "1",
		Subtotal: "100.00", TaxTotal: "21.00", TotalAmount: "121.00", IdempotencyKey: "invoice-10",
		Items: []SupplierInvoiceItem{{LineNumber: 1, Description: "Corks", Quantity: "10",
			Unit: "unit", UnitPrice: "10", TaxRate: "21", NetAmount: "100", TaxAmount: "21", TotalAmount: "121"}},
	})
	require.NoError(t, err)
	require.Equal(t, SupplierInvoiceDraft, invoice.Status)
	require.NoError(t, invoice.Issue())
	require.Equal(t, SupplierInvoiceIssued, invoice.Status)
	require.NoError(t, invoice.Issue())
	require.NoError(t, invoice.Void())
	require.Equal(t, SupplierInvoiceCancelled, invoice.Status)
	require.NoError(t, invoice.Void())
}

func TestSupplierPaymentRejectsOverAllocation(t *testing.T) {
	_, err := NewSupplierPayment(NewSupplierPaymentParams{
		SupplierID: "supplier", PaymentNumber: 1, PaymentDate: "2026-06-09", Currency: CurrencyARS,
		Amount: "100", PaymentMethod: "TRANSFER", IdempotencyKey: "payment-1",
		Allocations: []SupplierPaymentAllocation{{SupplierInvoiceID: "invoice", AllocatedAmount: "100.0001"}},
	})
	require.ErrorContains(t, err, "must not exceed")
}

func TestDocumentRejectsDecimalThatDatabaseWouldRound(t *testing.T) {
	_, err := NewSupplierInvoice(NewSupplierInvoiceParams{
		SupplierID: "supplier", DocumentType: SupplierInvoiceTypeInvoice, PointOfSale: 1,
		DocumentNumber: 10, IssueDate: "2026-06-09", Currency: CurrencyARS, ExchangeRate: "1.0000001",
		Subtotal: "100", TaxTotal: "21", TotalAmount: "121", IdempotencyKey: "invoice-scale",
		Items: []SupplierInvoiceItem{{LineNumber: 1, Description: "Corks", Quantity: "10",
			Unit: "unit", UnitPrice: "10", TaxRate: "21", NetAmount: "100", TaxAmount: "21", TotalAmount: "121"}},
	})
	require.ErrorContains(t, err, "exchange_rate")
}

func TestMonthlyVATPositionCalculatesTotalsByExactDecimal(t *testing.T) {
	position, err := NewMonthlyVATPosition("2026-06", []VATPositionLine{
		{TaxRate: "21.0000", SalesDebit: "210.10", PurchaseCredit: "100.05"},
		{TaxRate: "10.5", SalesDebit: "10.50", PurchaseCredit: "20.50"},
	})
	require.NoError(t, err)
	require.Equal(t, "220.6", position.SalesDebitTotal)
	require.Equal(t, "120.55", position.PurchaseCreditTotal)
	require.Equal(t, "100.05", position.NetVAT)
	require.Equal(t, "-10", position.Lines[1].NetVAT)
}

func TestMonthlyVATPositionAllowsCreditNotesToMakeAComponentNegative(t *testing.T) {
	position, err := NewMonthlyVATPosition("2026-06", []VATPositionLine{
		{TaxRate: "21", SalesDebit: "-5.50", PurchaseCredit: "10"},
	})
	require.NoError(t, err)
	require.Equal(t, "-15.5", position.NetVAT)
}

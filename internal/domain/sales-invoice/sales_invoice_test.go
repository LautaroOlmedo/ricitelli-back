package sales_invoice_test

import (
	"errors"
	"testing"

	sales_invoice "ricitelli-back/internal/domain/sales-invoice"
)

func validParams() sales_invoice.NewSalesInvoiceParams {
	return sales_invoice.NewSalesInvoiceParams{
		CustomerID: "customer-1", SaleOrderID: "order-1", DocumentType: sales_invoice.DocumentTypeInvoice,
		PointOfSale: 1, DocumentNumber: 10, IssueDate: "2026-06-09", DueDate: "2026-07-09",
		Currency: "ARS", ExchangeRate: "1.000000", Subtotal: "100.1000", TaxTotal: "21.0210",
		TotalAmount: "121.1210", IdempotencyKey: "invoice-key",
		Items: []sales_invoice.Item{{
			LineNumber: 1, ProductID: "product-1", Description: "Wine", Quantity: "1.0000",
			UnitPrice: "100.1000", TaxRate: "21.0000", NetAmount: "100.1000", TaxAmount: "21.0210", TotalAmount: "121.1210",
		}},
	}
}

func TestSalesInvoiceRequiresSaleOrderAndExactTotals(t *testing.T) {
	params := validParams()
	params.SaleOrderID = ""
	if _, err := sales_invoice.NewSalesInvoice(params); !errors.Is(err, sales_invoice.ErrInvalidSalesInvoice) {
		t.Fatalf("NewSalesInvoice() error = %v", err)
	}
	params = validParams()
	params.TotalAmount = "121.1211"
	if _, err := sales_invoice.NewSalesInvoice(params); !errors.Is(err, sales_invoice.ErrInvalidSalesInvoice) {
		t.Fatalf("NewSalesInvoice() error = %v", err)
	}
}

func TestSalesInvoiceIssueIsIdempotentAndMakesInvoiceImmutable(t *testing.T) {
	invoice, err := sales_invoice.NewSalesInvoice(validParams())
	if err != nil {
		t.Fatal(err)
	}
	if err := invoice.Issue(); err != nil {
		t.Fatal(err)
	}
	if err := invoice.Issue(); err != nil {
		t.Fatalf("second Issue() error = %v", err)
	}
	if invoice.GetStatus() != sales_invoice.StatusIssued || invoice.IsEditable() {
		t.Fatalf("issued invoice state = %s, editable=%v", invoice.GetStatus(), invoice.IsEditable())
	}
}

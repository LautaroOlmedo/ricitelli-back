package sales_administration

import (
	"context"
	"fmt"
	"testing"

	customer_receipt "ricitelli-back/internal/domain/customer-receipt"
	"ricitelli-back/internal/domain/remittance"
	sales_invoice "ricitelli-back/internal/domain/sales-invoice"
)

func TestIssueSalesInvoiceHasNoStockDependency(t *testing.T) {
	storage := &fakeStorage{}
	service := NewService(storage)
	if _, err := service.IssueSalesInvoice(context.Background(), "invoice-1"); err != nil {
		t.Fatal(err)
	}
	if storage.issueCalls != 1 {
		t.Fatalf("issue calls = %d, want 1", storage.issueCalls)
	}
}

func TestConfirmRemittanceDelegatesIdempotentDispatchUsingRemittanceID(t *testing.T) {
	storage := &fakeStorage{}
	service := NewService(storage)
	for range 2 {
		document, err := service.ConfirmRemittance(context.Background(), "remittance-1")
		if err != nil {
			t.Fatal(err)
		}
		if document.GetID() != "remittance-1" {
			t.Fatalf("remittance ID = %s", document.GetID())
		}
	}
	if storage.dispatches != 1 {
		t.Fatalf("physical dispatches = %d, want 1", storage.dispatches)
	}
	if storage.dispatchReference != "remittance-1" {
		t.Fatalf("dispatch reference = %s, want remittance-1", storage.dispatchReference)
	}
}

func TestOutstandingBalanceReturnsExactDecimalString(t *testing.T) {
	service := NewService(&fakeStorage{outstanding: "66.6667"})
	got, err := service.GetSalesInvoiceOutstandingBalance(context.Background(), "invoice-1")
	if err != nil || got != "66.6667" {
		t.Fatalf("balance = %q, %v", got, err)
	}
}

type fakeStorage struct {
	issueCalls        int
	dispatches        int
	dispatchReference string
	confirmed         bool
	outstanding       string
}

func (f *fakeStorage) SaveSalesInvoice(context.Context, sales_invoice.SalesInvoice) (*sales_invoice.SalesInvoice, error) {
	return nil, fmt.Errorf("unexpected SaveSalesInvoice")
}
func (f *fakeStorage) GetSalesInvoiceByID(context.Context, string) (*sales_invoice.SalesInvoice, error) {
	return nil, fmt.Errorf("unexpected GetSalesInvoiceByID")
}
func (f *fakeStorage) ListSalesInvoices(context.Context, SalesInvoiceFilter) ([]sales_invoice.SalesInvoice, error) {
	return nil, fmt.Errorf("unexpected ListSalesInvoices")
}
func (f *fakeStorage) IssueSalesInvoice(context.Context, string) (*sales_invoice.SalesInvoice, error) {
	f.issueCalls++
	return &sales_invoice.SalesInvoice{}, nil
}
func (f *fakeStorage) GetSalesInvoiceOutstandingBalance(context.Context, string) (string, error) {
	return f.outstanding, nil
}
func (f *fakeStorage) SaveRemittance(context.Context, remittance.Remittance) (*remittance.Remittance, error) {
	return nil, fmt.Errorf("unexpected SaveRemittance")
}
func (f *fakeStorage) GetRemittanceByID(context.Context, string) (*remittance.Remittance, error) {
	return nil, fmt.Errorf("unexpected GetRemittanceByID")
}
func (f *fakeStorage) ListRemittances(context.Context, RemittanceFilter) ([]remittance.Remittance, error) {
	return nil, fmt.Errorf("unexpected ListRemittances")
}
func (f *fakeStorage) ConfirmRemittance(_ context.Context, id, _ string) (*remittance.Remittance, error) {
	if !f.confirmed {
		f.confirmed = true
		f.dispatches++
		f.dispatchReference = id
	}
	document := remittance.ReconstituteRemittance(id, remittance.NewRemittanceParams{
		CustomerID: "customer-1", SaleOrderID: "order-1", PointOfSale: 1, DocumentNumber: 1,
		IssueDate: "2026-06-09", IdempotencyKey: "key",
		Items: []remittance.Item{{LineNumber: 1, ProductID: "product-1", Description: "Wine", Quantity: "1"}},
	}, remittance.StatusConfirmed, "", "")
	return &document, nil
}
func (f *fakeStorage) SaveCustomerReceipt(context.Context, customer_receipt.CustomerReceipt) (*customer_receipt.CustomerReceipt, error) {
	return nil, fmt.Errorf("unexpected SaveCustomerReceipt")
}
func (f *fakeStorage) GetCustomerReceiptByID(context.Context, string) (*customer_receipt.CustomerReceipt, error) {
	return nil, fmt.Errorf("unexpected GetCustomerReceiptByID")
}
func (f *fakeStorage) ListCustomerReceipts(context.Context, CustomerReceiptFilter) ([]customer_receipt.CustomerReceipt, error) {
	return nil, fmt.Errorf("unexpected ListCustomerReceipts")
}
func (f *fakeStorage) PostCustomerReceipt(context.Context, string) (*customer_receipt.CustomerReceipt, error) {
	return nil, fmt.Errorf("unexpected PostCustomerReceipt")
}

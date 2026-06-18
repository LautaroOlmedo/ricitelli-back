package sales_administration

import (
	"context"

	"ricitelli-back/internal/auth"
	customer_receipt "ricitelli-back/internal/domain/customer-receipt"
	"ricitelli-back/internal/domain/remittance"
	sales_invoice "ricitelli-back/internal/domain/sales-invoice"
)

// Storage is implemented by the PostgreSQL repository in dedicated V1A files.
// Confirmation/posting methods own their DB transaction and are idempotent.
type Storage interface {
	SaveSalesInvoice(context.Context, sales_invoice.SalesInvoice) (*sales_invoice.SalesInvoice, error)
	GetSalesInvoiceByID(context.Context, string) (*sales_invoice.SalesInvoice, error)
	ListSalesInvoices(context.Context, SalesInvoiceFilter) ([]sales_invoice.SalesInvoice, error)
	IssueSalesInvoice(context.Context, string) (*sales_invoice.SalesInvoice, error)
	GetSalesInvoiceOutstandingBalance(context.Context, string) (string, error)

	SaveRemittance(context.Context, remittance.Remittance) (*remittance.Remittance, error)
	GetRemittanceByID(context.Context, string) (*remittance.Remittance, error)
	ListRemittances(context.Context, RemittanceFilter) ([]remittance.Remittance, error)
	ConfirmRemittance(context.Context, string, string) (*remittance.Remittance, error)

	SaveCustomerReceipt(context.Context, customer_receipt.CustomerReceipt) (*customer_receipt.CustomerReceipt, error)
	GetCustomerReceiptByID(context.Context, string) (*customer_receipt.CustomerReceipt, error)
	ListCustomerReceipts(context.Context, CustomerReceiptFilter) ([]customer_receipt.CustomerReceipt, error)
	PostCustomerReceipt(context.Context, string) (*customer_receipt.CustomerReceipt, error)
}

type SalesInvoiceFilter struct {
	CustomerID  string
	SaleOrderID string
	Status      string
}

type RemittanceFilter struct {
	CustomerID  string
	SaleOrderID string
	Status      string
}

type CustomerReceiptFilter struct {
	CustomerID  string
	SaleOrderID string
	Status      string
}

type Service struct {
	Storage Storage
}

func NewService(storage Storage) *Service {
	return &Service{Storage: storage}
}

func userIDFromContext(ctx context.Context) string {
	if claims, ok := auth.ClaimsFromContext(ctx); ok {
		return claims.UserID
	}
	return "system"
}

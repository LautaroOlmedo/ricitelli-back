package sales_administration

import (
	"context"

	sales_invoice "ricitelli-back/internal/domain/sales-invoice"
)

func (s *Service) CreateSalesInvoice(ctx context.Context, params sales_invoice.NewSalesInvoiceParams) (*sales_invoice.SalesInvoice, error) {
	invoice, err := sales_invoice.NewSalesInvoice(params)
	if err != nil {
		return nil, err
	}
	return s.Storage.SaveSalesInvoice(ctx, invoice)
}

func (s *Service) GetSalesInvoiceByID(ctx context.Context, id string) (*sales_invoice.SalesInvoice, error) {
	return s.Storage.GetSalesInvoiceByID(ctx, id)
}

func (s *Service) ListSalesInvoices(ctx context.Context, filter SalesInvoiceFilter) ([]sales_invoice.SalesInvoice, error) {
	return s.Storage.ListSalesInvoices(ctx, filter)
}

// IssueSalesInvoice has no inventory dependency or stock side effect.
func (s *Service) IssueSalesInvoice(ctx context.Context, id string) (*sales_invoice.SalesInvoice, error) {
	return s.Storage.IssueSalesInvoice(ctx, id)
}

func (s *Service) GetSalesInvoiceOutstandingBalance(ctx context.Context, id string) (string, error) {
	return s.Storage.GetSalesInvoiceOutstandingBalance(ctx, id)
}

package purchasing_administration

import (
	"context"

	domain "ricitelli-back/internal/domain/purchasing-administration"
)

func (s *Service) CreateSupplierInvoice(ctx context.Context, params domain.NewSupplierInvoiceParams) (domain.SupplierInvoice, error) {
	return s.Storage.CreateSupplierInvoice(ctx, params)
}
func (s *Service) GetSupplierInvoiceByID(ctx context.Context, id string) (*domain.SupplierInvoice, error) {
	return s.Storage.GetSupplierInvoiceByID(ctx, id)
}
func (s *Service) GetSupplierInvoices(ctx context.Context, supplierID string) ([]domain.SupplierInvoice, error) {
	return s.Storage.GetSupplierInvoices(ctx, supplierID)
}
func (s *Service) IssueSupplierInvoice(ctx context.Context, id string) (*domain.SupplierInvoice, error) {
	invoice, err := s.Storage.GetSupplierInvoiceByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if err := invoice.Issue(); err != nil {
		return nil, err
	}
	return s.Storage.UpdateSupplierInvoiceStatus(ctx, id, invoice.Status)
}
func (s *Service) VoidSupplierInvoice(ctx context.Context, id string) (*domain.SupplierInvoice, error) {
	invoice, err := s.Storage.GetSupplierInvoiceByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if err := invoice.Void(); err != nil {
		return nil, err
	}
	return s.Storage.UpdateSupplierInvoiceStatus(ctx, id, invoice.Status)
}

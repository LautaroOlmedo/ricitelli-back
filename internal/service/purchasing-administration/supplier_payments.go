package purchasing_administration

import (
	"context"

	domain "ricitelli-back/internal/domain/purchasing-administration"
)

func (s *Service) CreateSupplierPayment(ctx context.Context, params domain.NewSupplierPaymentParams) (domain.SupplierPayment, error) {
	return s.Storage.CreateSupplierPayment(ctx, params)
}
func (s *Service) GetSupplierPaymentByID(ctx context.Context, id string) (*domain.SupplierPayment, error) {
	return s.Storage.GetSupplierPaymentByID(ctx, id)
}
func (s *Service) GetSupplierPayments(ctx context.Context, supplierID string) ([]domain.SupplierPayment, error) {
	return s.Storage.GetSupplierPayments(ctx, supplierID)
}
func (s *Service) PostSupplierPayment(ctx context.Context, id string) (*domain.SupplierPayment, error) {
	payment, err := s.Storage.GetSupplierPaymentByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if err := payment.Post(); err != nil {
		return nil, err
	}
	return s.Storage.UpdateSupplierPaymentStatus(ctx, id, payment.Status)
}
func (s *Service) VoidSupplierPayment(ctx context.Context, id string) (*domain.SupplierPayment, error) {
	payment, err := s.Storage.GetSupplierPaymentByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if err := payment.Void(); err != nil {
		return nil, err
	}
	return s.Storage.UpdateSupplierPaymentStatus(ctx, id, payment.Status)
}
func (s *Service) GetSupplierOutstandingBalance(ctx context.Context, supplierID string, currency domain.Currency) (*domain.SupplierOutstandingBalance, error) {
	return s.Storage.GetSupplierOutstandingBalance(ctx, supplierID, currency)
}

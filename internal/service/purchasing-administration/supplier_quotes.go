package purchasing_administration

import (
	"context"

	domain "ricitelli-back/internal/domain/purchasing-administration"
)

func (s *Service) CreateSupplierQuote(ctx context.Context, params domain.NewSupplierQuoteParams) (domain.SupplierQuote, error) {
	return s.Storage.CreateSupplierQuote(ctx, params)
}
func (s *Service) GetSupplierQuoteByID(ctx context.Context, id string) (*domain.SupplierQuote, error) {
	return s.Storage.GetSupplierQuoteByID(ctx, id)
}
func (s *Service) GetSupplierQuotes(ctx context.Context, supplierID string) ([]domain.SupplierQuote, error) {
	return s.Storage.GetSupplierQuotes(ctx, supplierID)
}
func (s *Service) UpdateSupplierQuoteStatus(ctx context.Context, id string, status domain.SupplierQuoteStatus) (*domain.SupplierQuote, error) {
	quote, err := s.Storage.GetSupplierQuoteByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if err := quote.UpdateStatus(status); err != nil {
		return nil, err
	}
	return s.Storage.UpdateSupplierQuoteStatus(ctx, id, status)
}

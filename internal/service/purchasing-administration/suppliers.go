package purchasing_administration

import (
	"context"

	domain "ricitelli-back/internal/domain/purchasing-administration"
)

func (s *Service) CreateSupplier(ctx context.Context, params domain.NewSupplierParams) (domain.Supplier, error) {
	return s.Storage.CreateSupplier(ctx, params)
}

func (s *Service) GetSupplierByID(ctx context.Context, id string) (*domain.Supplier, error) {
	return s.Storage.GetSupplierByID(ctx, id)
}

func (s *Service) GetSuppliers(ctx context.Context, includeInactive bool) ([]domain.Supplier, error) {
	return s.Storage.GetSuppliers(ctx, includeInactive)
}

func (s *Service) UpdateSupplier(ctx context.Context, id string, params domain.UpdateSupplierParams) (*domain.Supplier, error) {
	return s.Storage.UpdateSupplier(ctx, id, params)
}

func (s *Service) DeactivateSupplier(ctx context.Context, id string) (*domain.Supplier, error) {
	supplier, err := s.Storage.GetSupplierByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if err := supplier.Deactivate(); err != nil {
		return nil, err
	}
	return s.Storage.DeactivateSupplier(ctx, id)
}

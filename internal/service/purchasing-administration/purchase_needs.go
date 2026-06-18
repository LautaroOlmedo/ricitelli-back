package purchasing_administration

import (
	"context"

	domain "ricitelli-back/internal/domain/purchasing-administration"
)

func (s *Service) CreatePurchaseNeed(ctx context.Context, params domain.NewPurchaseNeedParams) (domain.PurchaseNeed, error) {
	return s.Storage.CreatePurchaseNeed(ctx, params)
}
func (s *Service) GetPurchaseNeedByID(ctx context.Context, id string) (*domain.PurchaseNeed, error) {
	return s.Storage.GetPurchaseNeedByID(ctx, id)
}
func (s *Service) GetPurchaseNeeds(ctx context.Context) ([]domain.PurchaseNeed, error) {
	return s.Storage.GetPurchaseNeeds(ctx)
}
func (s *Service) UpdatePurchaseNeedStatus(ctx context.Context, id string, status domain.PurchaseNeedStatus) (*domain.PurchaseNeed, error) {
	need, err := s.Storage.GetPurchaseNeedByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if err := need.UpdateStatus(status); err != nil {
		return nil, err
	}
	return s.Storage.UpdatePurchaseNeedStatus(ctx, id, status)
}

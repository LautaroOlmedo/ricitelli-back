package purchasing_administration

import (
	"context"

	domain "ricitelli-back/internal/domain/purchasing-administration"
)

func (s *Service) GetMonthlyVATPosition(ctx context.Context, month string) (*domain.MonthlyVATPosition, error) {
	if _, err := domain.NewMonthlyVATPosition(month, nil); err != nil {
		return nil, err
	}
	return s.Storage.GetMonthlyVATPosition(ctx, month)
}

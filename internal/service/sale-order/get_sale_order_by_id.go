package sale_order

import (
	"context"
	sale_order "ricitelli-back/internal/domain/sale-order"
)

func (s *Service) GetSaleOrderByID(ctx context.Context, id string) (*sale_order.SaleOrder, error) {
	order, err := s.Storage.GetSaleOrderByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if err := s.attachAdministrativeSummary(ctx, order); err != nil {
		return nil, err
	}
	return order, nil
}

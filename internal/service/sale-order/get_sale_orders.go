package sale_order

import (
	"context"
	sale_order "ricitelli-back/internal/domain/sale-order"
)

func (s *Service) GetSaleOrders(ctx context.Context) ([]sale_order.SaleOrder, error) {
	orders, err := s.Storage.GetSaleOrders(ctx)
	if err != nil {
		return nil, err
	}
	if err := s.attachAdministrativeSummaries(ctx, orders); err != nil {
		return nil, err
	}
	return orders, nil
}

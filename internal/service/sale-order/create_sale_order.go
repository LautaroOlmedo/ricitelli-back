package sale_order

import (
	"context"

	sale_order "ricitelli-back/internal/domain/sale-order"
)

func (s *Service) CreateSaleOrder(ctx context.Context, params sale_order.NewSaleOrderParams) (sale_order.SaleOrder, error) {
	order, err := s.Storage.CreateSaleOrder(ctx, params)
	if err != nil {
		return sale_order.SaleOrder{}, err
	}
	if err := s.attachAdministrativeSummary(ctx, &order); err != nil {
		return sale_order.SaleOrder{}, err
	}
	return order, nil
}

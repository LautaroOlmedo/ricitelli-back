package sale_order

import (
	"context"

	sale_order "ricitelli-back/internal/domain/sale-order"
)

func (s *Service) CreateSaleOrder(ctx context.Context, params sale_order.NewSaleOrderParams) (sale_order.SaleOrder, error) {
	return s.Storage.CreateSaleOrder(ctx, params)
}

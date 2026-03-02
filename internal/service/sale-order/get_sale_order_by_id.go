package sale_order

import (
	"context"
	sale_order "ricitelli-back/internal/domain/sale-order"
)

func (s *Service) GetSaleOrderByID(ctx context.Context, id string) (*sale_order.SaleOrder, error) {
	return s.Storage.GetSaleOrderByID(ctx, id)
}

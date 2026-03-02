package sale_order

import (
	"context"
	sale_order "ricitelli-back/internal/domain/sale-order"
)

func (s *Service) GetSaleOrders(ctx context.Context) ([]sale_order.SaleOrder, error) {
	return s.Storage.GetSaleOrders(ctx)
}

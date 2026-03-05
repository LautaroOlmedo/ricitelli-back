package sale_order

import (
	"context"

	sale_order "ricitelli-back/internal/domain/sale-order"
)

func (s *Service) UpdateSaleOrderStatus(ctx context.Context, id string, status sale_order.Status) (*sale_order.SaleOrder, error) {
	return s.Storage.UpdateSaleOrderStatus(ctx, id, status)
}

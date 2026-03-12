package production_order

import (
	"context"
	"errors"

	production_order "ricitelli-back/internal/domain/production-order"
)

func (s *Service) GetProductionOrdersBySaleOrder(ctx context.Context, saleOrderID string) ([]production_order.ProductionOrder, error) {
	if saleOrderID == "" {
		return nil, errors.New("sale_order_id cannot be empty")
	}
	return s.Storage.GetProductionOrdersBySaleOrder(ctx, saleOrderID)
}

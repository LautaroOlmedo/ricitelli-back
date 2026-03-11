package production_order

import (
	"context"
	production_order "ricitelli-back/internal/domain/production-order"
)

func (s *Service) UpdateProductionOrderStatus(ctx context.Context, id string, newStatus production_order.Status) (*production_order.ProductionOrder, error) {
	return s.Storage.UpdateProductionOrderStatus(ctx, id, newStatus)
}

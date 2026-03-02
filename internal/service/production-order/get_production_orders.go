package production_order

import (
	"context"
	production_order "ricitelli-back/internal/domain/production-order"
)

func (s *Service) GetProductionOrders(ctx context.Context) ([]production_order.ProductionOrder, error) {
	return s.Storage.GetProductionOrders(ctx)
}

package production_order

import (
	"context"
	production_order "ricitelli-back/internal/domain/production-order"
)

func (s *Service) GetProductionOrderByID(ctx context.Context, id string) (*production_order.ProductionOrder, error) {
	return s.Storage.GetProductionOrderByID(ctx, id)
}

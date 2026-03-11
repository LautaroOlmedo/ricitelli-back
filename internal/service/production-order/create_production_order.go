package production_order

import (
	"context"
	production_order "ricitelli-back/internal/domain/production-order"
	"ricitelli-back/internal/entities"
)

func (s *Service) CreateProductionOrder(ctx context.Context, salesOrderID string, items []entities.ProductionItem) (production_order.ProductionOrder, error) {
	return s.Storage.CreateProductionOrder(ctx, salesOrderID, items)
}

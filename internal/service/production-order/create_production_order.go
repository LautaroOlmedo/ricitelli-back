package production_order

import (
	"context"
	"ricitelli-back/internal/entities"
)

func (s *Service) CreateProductionOrder(ctx context.Context, salesOrderID string, items []entities.ProductionItem) error {
	return s.Storage.CreateProductionOrder(ctx, salesOrderID, items)
}

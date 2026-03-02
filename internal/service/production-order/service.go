package production_order

import (
	"context"
	production_order "ricitelli-back/internal/domain/production-order"
	"ricitelli-back/internal/entities"
)

type ProductionOrderStorage interface {
	CreateProductionOrder(ctx context.Context, salesOrderID string, items []entities.ProductionItem) error
	GetProductionOrderByID(ctx context.Context, id string) (*production_order.ProductionOrder, error)
	GetProductionOrders(ctx context.Context) ([]production_order.ProductionOrder, error)
}

type Service struct {
	Storage ProductionOrderStorage
}

func NewProductionOrderService(productionOrderStorage ProductionOrderStorage) *Service {
	return &Service{
		Storage: productionOrderStorage,
	}
}

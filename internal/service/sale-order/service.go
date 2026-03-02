package sale_order

import (
	"context"
	sale_order "ricitelli-back/internal/domain/sale-order"
	valueObject "ricitelli-back/internal/value-object"
)

type SaleOrderStorage interface {
	CreateSaleOrder(ctx context.Context, customerID string, items []valueObject.SaleOrderItem) (sale_order.SaleOrder, error)
	GetSaleOrderByID(ctx context.Context, id string) (*sale_order.SaleOrder, error)
	GetSaleOrders(ctx context.Context) ([]sale_order.SaleOrder, error)
}

type Service struct {
	Storage SaleOrderStorage
}

func NewSaleOrderService(saleOrderStorage SaleOrderStorage) *Service {
	return &Service{
		Storage: saleOrderStorage,
	}
}

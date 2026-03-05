package sale_order

import (
	"context"

	sale_order "ricitelli-back/internal/domain/sale-order"
)

type SaleOrderStorage interface {
	CreateSaleOrder(ctx context.Context, params sale_order.NewSaleOrderParams) (sale_order.SaleOrder, error)
	GetSaleOrderByID(ctx context.Context, id string) (*sale_order.SaleOrder, error)
	GetSaleOrders(ctx context.Context) ([]sale_order.SaleOrder, error)
	UpdateSaleOrderStatus(ctx context.Context, id string, status sale_order.Status) (*sale_order.SaleOrder, error)
}

type Service struct {
	Storage SaleOrderStorage
}

func NewSaleOrderService(saleOrderStorage SaleOrderStorage) *Service {
	return &Service{Storage: saleOrderStorage}
}

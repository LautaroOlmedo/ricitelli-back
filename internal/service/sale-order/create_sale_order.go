package sale_order

import (
	"context"
	sale_order "ricitelli-back/internal/domain/sale-order"
	valueObject "ricitelli-back/internal/value-object"
)

func (s *Service) CreateSaleOrder(ctx context.Context, customerID string, items []valueObject.SaleOrderItem) (sale_order.SaleOrder, error) {
	return s.Storage.CreateSaleOrder(ctx, customerID, items)
}

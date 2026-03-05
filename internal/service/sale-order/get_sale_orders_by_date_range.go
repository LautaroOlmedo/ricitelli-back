package sale_order

import (
	"context"
	"errors"

	sale_order "ricitelli-back/internal/domain/sale-order"
)

func (s *Service) GetSaleOrdersByDateRange(ctx context.Context, from, to string) ([]sale_order.SaleOrder, error) {
	if from == "" || to == "" {
		return nil, errors.New("from_date and to_date are required")
	}
	return s.Storage.GetSaleOrdersByDateRange(ctx, from, to)
}

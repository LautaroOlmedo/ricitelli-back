package sale_order

import (
	"context"

	sale_order "ricitelli-back/internal/domain/sale-order"
)

func (s *Service) UpdateSaleOrderStatus(ctx context.Context, id string, status sale_order.Status) (*sale_order.SaleOrder, error) {
	order, err := s.Storage.UpdateSaleOrderStatus(ctx, id, status)
	if err != nil {
		return nil, err
	}
	if err := s.attachAdministrativeSummary(ctx, order); err != nil {
		return nil, err
	}
	return order, nil
}

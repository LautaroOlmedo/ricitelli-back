package customer

import (
	"context"

	customer_domain "ricitelli-back/internal/domain/customer"
)

func (s *Service) SearchCustomersBySocialReason(ctx context.Context, query string) ([]customer_domain.Customer, error) {
	return s.Storage.SearchCustomersBySocialReason(ctx, query)
}

package customer

import (
	"context"

	customer_domain "ricitelli-back/internal/domain/customer"
)

func (s *Service) GetCustomers(ctx context.Context) ([]customer_domain.Customer, error) {
	return s.Storage.GetCustomers(ctx)
}

package customer

import (
	"context"

	customer_domain "ricitelli-back/internal/domain/customer"
)

func (s *Service) CreateCustomer(ctx context.Context, params customer_domain.NewCustomerParams) (customer_domain.Customer, error) {
	return s.Storage.CreateCustomer(ctx, params)
}

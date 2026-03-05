package customer

import (
	"context"

	customer_domain "ricitelli-back/internal/domain/customer"
)

func (s *Service) GetCustomerByID(ctx context.Context, id string) (*customer_domain.Customer, error) {
	return s.Storage.GetCustomerByID(ctx, id)
}

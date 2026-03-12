package customer

import (
	"context"
	"errors"

	customer_domain "ricitelli-back/internal/domain/customer"
)

func (s *Service) UpdateCustomer(ctx context.Context, id string, params customer_domain.UpdateCustomerParams) (*customer_domain.Customer, error) {
	if id == "" {
		return nil, errors.New("id cannot be empty")
	}
	return s.Storage.UpdateCustomer(ctx, id, params)
}

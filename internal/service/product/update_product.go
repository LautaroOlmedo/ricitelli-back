package product

import (
	"context"
	"errors"

	valueObject "ricitelli-back/internal/value-object"
)

func (s *ProductService) UpdateProduct(ctx context.Context, id, name string, bods []valueObject.BillOfDrySupply) error {
	if id == "" {
		return errors.New("id cannot be empty")
	}
	if name == "" {
		return errors.New("name cannot be empty")
	}
	return s.Storage.UpdateProduct(ctx, id, name, bods)
}


package product

import (
	"context"
	valueObject "ricitelli-back/internal/value-object"
)

func (s *ProductService) CreateProduct(ctx context.Context, name string, bods []valueObject.BillOfDrySupply) error {
	return s.Storage.CreateProduct(ctx, name, bods)
}

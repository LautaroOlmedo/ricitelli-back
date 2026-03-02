package product

import (
	"context"
	valueObject "ricitelli-back/internal/value-object"
)

func (s *Service) CreateProduct(ctx context.Context, name string, bods []valueObject.BillOfDrySupply) error {
	return s.productStorage.CreateProduct(ctx, name, bods)
}

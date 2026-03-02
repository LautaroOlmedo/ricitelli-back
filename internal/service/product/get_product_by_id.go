package product

import (
	"context"
	"ricitelli-back/internal/domain/product"
)

func (s *Service) GetProductByID(ctx context.Context, id string) (*product.Product, error) {
	return s.productStorage.GetProductByID(ctx, id)
}

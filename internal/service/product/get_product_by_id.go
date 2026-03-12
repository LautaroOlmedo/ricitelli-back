package product

import (
	"context"
	"ricitelli-back/internal/domain/product"
)

func (s *ProductService) GetProductByID(ctx context.Context, id string) (*product.Product, error) {
	return s.Storage.GetProductByID(ctx, id)
}

package product

import (
	"context"
	"ricitelli-back/internal/domain/product"
)

func (s *ProductService) GetProducts(ctx context.Context) ([]product.Product, error) {
	return s.Storage.GetProducts(ctx)
}

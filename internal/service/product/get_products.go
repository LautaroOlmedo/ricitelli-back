package product

import (
	"context"
	"ricitelli-back/internal/domain/product"
)

func (s *Service) GetProducts(ctx context.Context) ([]product.Product, error) {
	return s.productStorage.GetProducts(ctx)
}

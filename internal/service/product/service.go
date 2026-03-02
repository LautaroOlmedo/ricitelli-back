package product

import (
	"context"
	"ricitelli-back/internal/domain/product"
	valueObject "ricitelli-back/internal/value-object"
)

type ProductStorage interface {
	CreateProduct(ctx context.Context, name string, bods []valueObject.BillOfDrySupply) error
	GetProductByID(ctx context.Context, id string) (*product.Product, error)
	GetProducts(ctx context.Context) ([]product.Product, error)
}
type Service struct {
	productStorage ProductStorage
}

func NewProductService(productStorage ProductStorage) *Service {
	return &Service{productStorage: productStorage}
}

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
	UpdateProduct(ctx context.Context, id, name string, bods []valueObject.BillOfDrySupply) error
}

// ProductService is the concrete service type (renamed from Service to avoid collision with other packages).
type ProductService struct {
	Storage ProductStorage
}

// Deprecated alias kept for backward compatibility with existing code.
type Service = ProductService

func NewProductService(productStorage ProductStorage) *ProductService {
	return &ProductService{Storage: productStorage}
}

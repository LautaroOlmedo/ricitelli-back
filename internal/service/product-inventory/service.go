package product_inventory

import (
	product_inventory "ricitelli-back/internal/domain/product-inventory"
)

type ProductInventoryStorage interface {
	CreateProductInventory(prodInventory product_inventory.ProductInventory) error
	GetProductInventory(productID string) (*product_inventory.ProductInventory, error)
}

type Service struct {
	Storage ProductInventoryStorage
}

func NewProductInventoryService(storage ProductInventoryStorage) *Service {
	return &Service{Storage: storage}
}

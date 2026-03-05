package product_inventory

import (
	product_inventory "ricitelli-back/internal/domain/product-inventory"
)

func (s *Service) SaveProductInventory(inv product_inventory.ProductInventory) error {
	return s.Storage.SaveProductInventory(inv)
}

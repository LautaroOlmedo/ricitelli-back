package product_inventory

import product_inventory "ricitelli-back/internal/domain/product-inventory"

func (s *Service) GetProductInventory(productID string) (*product_inventory.ProductInventory, error) {
	return s.Storage.GetProductInventory(productID)
}

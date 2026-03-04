package product_inventory

import product_inventory "ricitelli-back/internal/domain/product-inventory"

func (s *Service) CreateProductInventory(productID, sku string) error {
	newProdInventory, err := product_inventory.NewProductInventory(productID, sku)
	if err != nil {
		return err
	}
	return s.Storage.CreateProductInventory(newProdInventory)
}

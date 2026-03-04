package application_service

import (
	"context"
	"fmt"
	"ricitelli-back/internal/entities"
	valueObject "ricitelli-back/internal/value-object"
)

func (s *Service) CreateOrder(
	ctx context.Context,
	customerID string,
	items []valueObject.SaleOrderItem,
) error {

	saleOrder, err := s.SaleOrderService.CreateSaleOrder(ctx, customerID, items)
	if err != nil {
		fmt.Println(err)
		return err
	}

	var productionItems []entities.ProductionItem

	for _, item := range saleOrder.GetItems() {

		product, err := s.ProductService.GetProductByID(ctx, item.ProductID)
		if err != nil {
			fmt.Println(err)
			return err
		}
		/*productInventory, err := s.ProductInventoryService.GetProductInventory(product.GetID())
		if productInventory == nil {
			fmt.Println("Not Product Inventory")
			err = s.ProductInventoryService.CreateProductInventory(product.GetID(), product.GetID())
			if err != nil {
				return err
			}
		}
		fmt.Println("Product Inventory", productInventory)
		if err != nil {
			return err
		}
		fmt.Println("product inventory:", productInventory)
		if productInventory.AvailableUndressed() >= int64(item.Quantity) {
			err = productInventory.Reserve(saleOrder.GetID(), item.Quantity)
			if err != nil {
				return err
			}
			fmt.Println("product inventory2:", productInventory)

		} else {
			fmt.Println("reserving")*/
		reqs := product.CalculateRequirements(item.Quantity)

		productionItems = append(productionItems, entities.ProductionItem{
			ProductID:    item.ProductID,
			Quantity:     item.Quantity,
			Requirements: reqs,
		})

	}

	err = s.ProductionOrderService.CreateProductionOrder(
		ctx,
		saleOrder.GetID(),
		productionItems,
	)
	if err != nil {
		fmt.Println(err)
		return err
	}

	return nil
}

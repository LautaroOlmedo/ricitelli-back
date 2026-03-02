package application_service

import (
	"context"
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
		return err
	}

	var productionItems []entities.ProductionItem

	for _, item := range saleOrder.GetItems() {

		product, err := s.ProductService.GetProductByID(ctx, item.ProductID)
		if err != nil {
			return err
		}

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
		return err
	}

	return nil
}

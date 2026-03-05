package application_service

import (
	"context"
	"fmt"

	sale_order "ricitelli-back/internal/domain/sale-order"
	"ricitelli-back/internal/entities"
	valueObject "ricitelli-back/internal/value-object"
)

// CreateOrderParams contains all fields needed to create a sale + production order atomically.
type CreateOrderParams struct {
	CustomerID         string
	Items              []OrderItem
	Currency           sale_order.Currency
	Market             sale_order.Market
	DestinationCountry string
	SaleType           sale_order.SaleType
}

type OrderItem struct {
	ProductID string
	Quantity  uint64
	UnitPrice float32
}

// CreateOrder atomically creates a SaleOrder + ProductionOrder and commits dry supply stock.
func (s *Service) CreateOrder(ctx context.Context, params CreateOrderParams) error {
	soItems := make([]valueObject.SaleOrderItem, 0, len(params.Items))
	for _, item := range params.Items {
		soItems = append(soItems, valueObject.SaleOrderItem{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
			UnitPrice: item.UnitPrice,
		})
	}

	soParams := sale_order.NewSaleOrderParams{
		CustomerID:         params.CustomerID,
		Items:              soItems,
		Currency:           params.Currency,
		Market:             params.Market,
		DestinationCountry: params.DestinationCountry,
		SaleType:           params.SaleType,
	}
	createdSaleOrder, err := s.SaleOrderService.CreateSaleOrder(ctx, soParams)
	if err != nil {
		fmt.Println("CreateOrder: CreateSaleOrder error:", err)
		return err
	}

	var productionItems []entities.ProductionItem
	for _, item := range createdSaleOrder.GetItems() {
		product, err := s.ProductService.GetProductByID(ctx, item.ProductID)
		if err != nil {
			fmt.Println("CreateOrder: GetProductByID error:", err)
			return err
		}
		reqs := product.CalculateRequirements(item.Quantity)
		productionItems = append(productionItems, entities.ProductionItem{
			ProductID:    item.ProductID,
			Quantity:     item.Quantity,
			Requirements: reqs,
		})
	}

	if err := s.ProductionOrderService.CreateProductionOrder(ctx, createdSaleOrder.GetID(), productionItems); err != nil {
		fmt.Println("CreateOrder: CreateProductionOrder error:", err)
		return err
	}

	// Commit dry supply stock for each material requirement (best-effort)
	for _, pi := range productionItems {
		for _, req := range pi.Requirements {
			if commitErr := s.DrySupplyService.CommitStock(ctx, req.DrySupplyID, req.Quantity, createdSaleOrder.GetID()); commitErr != nil {
				fmt.Printf("CreateOrder: CommitStock warning for %s: %v\n", req.DrySupplyID, commitErr)
			}
		}
	}

	return nil
}

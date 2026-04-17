package application_service

import (
	"context"
	"fmt"
	"time"

	production_order "ricitelli-back/internal/domain/production-order"
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

type itemStockInfo struct {
	productID      string
	requested      uint64
	needProduction uint64
}

// CreateOrder implements the full sale order flow:
//  1. Creates the sale order with status NEW.
//  2. If bottles need to be dressed: validates SV + dry supply stock, dresses bottles,
//     commits+consumes dry supplies, completes the production order.
//  3. The sale order always remains in status NEW after creation.
//     Status transitions (CONFIRMED → INVOICED → READY_TO_DISPATCH → DISPATCHED)
//     are driven manually by the user via the Kanban board.
func (s *Service) CreateOrder(ctx context.Context, params CreateOrderParams) error {
	soItems := make([]valueObject.SaleOrderItem, 0, len(params.Items))
	for _, item := range params.Items {
		soItems = append(soItems, valueObject.SaleOrderItem{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
			UnitPrice: item.UnitPrice,
		})
	}

	createdSaleOrder, err := s.SaleOrderService.CreateSaleOrder(ctx, sale_order.NewSaleOrderParams{
		CustomerID:         params.CustomerID,
		Items:              soItems,
		Currency:           params.Currency,
		Market:             params.Market,
		DestinationCountry: params.DestinationCountry,
		SaleType:           params.SaleType,
	})
	if err != nil {
		return fmt.Errorf("CreateOrder: create sale order: %w", err)
	}

	saleOrderID := createdSaleOrder.GetID()
	cancelOrder := func() { _ = s.cancelSaleOrder(ctx, saleOrderID) }

	// ── Step 1: assess whether any dressing is needed ─────────────────────────
	itemStocks, needsProduction, err := s.assessStock(ctx, createdSaleOrder.GetItems())
	if err != nil {
		cancelOrder()
		return fmt.Errorf("CreateOrder: assess stock: %w", err)
	}

	// ── Step 2: dress bottles + commit/consume dry supplies if needed ──────────
	if needsProduction {
		productionNeeds, err := s.buildAndValidateProductionNeeds(ctx, saleOrderID, itemStocks)
		if err != nil {
			cancelOrder()
			return err
		}

		productionItems := make([]entities.ProductionItem, 0, len(productionNeeds))
		for _, pn := range productionNeeds {
			productionItems = append(productionItems, entities.ProductionItem{
				ProductID:    pn.productID,
				Quantity:     pn.quantity,
				Requirements: pn.requirements,
			})
		}

		prodOrder, err := s.ProductionOrderService.CreateProductionOrder(ctx, saleOrderID, productionItems)
		if err != nil {
			cancelOrder()
			return fmt.Errorf("CreateOrder: create production order: %w", err)
		}

		if err := s.executeProduction(ctx, prodOrder.GetID(), saleOrderID, productionNeeds); err != nil {
			_, _ = s.ProductionOrderService.UpdateProductionOrderStatus(ctx, prodOrder.GetID(), production_order.StatusCancelled)
			cancelOrder()
			return err
		}

		if _, err := s.ProductionOrderService.UpdateProductionOrderStatus(ctx, prodOrder.GetID(), production_order.StatusCompleted); err != nil {
			return fmt.Errorf("CreateOrder: complete production order: %w", err)
		}
	}

	// Sale order stays in NEW — status progression is manual via Kanban.
	return nil
}

// ── helpers ──────────────────────────────────────────────────────────────────

// assessStock checks dressed availability per item and fills needProduction.
func (s *Service) assessStock(ctx context.Context, items []valueObject.SaleOrderItem) ([]itemStockInfo, bool, error) {
	result := make([]itemStockInfo, 0, len(items))
	needsProduction := false
	for _, item := range items {
		inv, err := s.ProductInventoryService.GetProductInventory(item.ProductID)
		if err != nil {
			return nil, false, fmt.Errorf("get inventory for product %s: %w", item.ProductID, err)
		}
		avail, _ := inv.AvailableDressed()
		var availU uint64
		if avail > 0 {
			availU = uint64(avail)
		}
		var needProd uint64
		if availU < item.Quantity {
			needProd = item.Quantity - availU
			needsProduction = true
		}
		result = append(result, itemStockInfo{
			productID:      item.ProductID,
			requested:      item.Quantity,
			needProduction: needProd,
		})
	}
	return result, needsProduction, nil
}

type productionNeed struct {
	productID    string
	quantity     uint64
	requirements []valueObject.MaterialRequirement
}

// buildAndValidateProductionNeeds pre-validates SV stock and dry supply availability
// before any mutations happen. Returns error (and expects caller to cancel the sale order).
func (s *Service) buildAndValidateProductionNeeds(ctx context.Context, saleOrderID string, itemStocks []itemStockInfo) ([]productionNeed, error) {
	needs := make([]productionNeed, 0)
	for _, is := range itemStocks {
		if is.needProduction == 0 {
			continue
		}
		product, err := s.ProductService.GetProductByID(ctx, is.productID)
		if err != nil {
			return nil, fmt.Errorf("get product %s: %w", is.productID, err)
		}
		inv, err := s.ProductInventoryService.GetProductInventory(is.productID)
		if err != nil {
			return nil, fmt.Errorf("get inventory for %s: %w", is.productID, err)
		}
		if inv.AvailableUndressed() < int64(is.needProduction) {
			return nil, fmt.Errorf("insufficient undressed stock for product %s: need %d, have %d",
				is.productID, is.needProduction, inv.AvailableUndressed())
		}
		reqs := product.CalculateRequirements(is.needProduction)
		for _, req := range reqs {
			dsInv, err := s.DrySupplyService.GetDrySupplyInventory(ctx, req.DrySupplyID)
			if err != nil {
				return nil, fmt.Errorf("get dry supply inventory %s: %w", req.DrySupplyID, err)
			}
			if dsInv.AvailableStock() < int64(req.Quantity) {
				return nil, fmt.Errorf("insufficient dry supply %s: need %d, have %d",
					req.DrySupplyID, req.Quantity, dsInv.AvailableStock())
			}
		}
		needs = append(needs, productionNeed{
			productID:    is.productID,
			quantity:     is.needProduction,
			requirements: reqs,
		})
	}
	return needs, nil
}

// executeProduction dresses bottles and commits+consumes dry supplies for each production need.
func (s *Service) executeProduction(ctx context.Context, prodOrderID, saleOrderID string, needs []productionNeed) error {
	lotBase := time.Now().Format("020106")
	for _, pn := range needs {
		inv, err := s.ProductInventoryService.GetProductInventory(pn.productID)
		if err != nil {
			return fmt.Errorf("get inventory for dressing %s: %w", pn.productID, err)
		}
		lotNumber := fmt.Sprintf("L-%s-%d-%s", lotBase, pn.quantity, prodOrderID[:8])
		if err := inv.ConvertSVtoPT(prodOrderID, pn.quantity, lotNumber, userIDFromCtx(ctx)); err != nil {
			return fmt.Errorf("dress bottles for %s: %w", pn.productID, err)
		}
		if err := s.ProductInventoryService.SaveProductInventory(*inv); err != nil {
			return fmt.Errorf("save inventory for %s: %w", pn.productID, err)
		}
		for _, req := range pn.requirements {
			if err := s.DrySupplyService.CommitStock(ctx, req.DrySupplyID, req.Quantity, prodOrderID); err != nil {
				return fmt.Errorf("commit dry supply %s: %w", req.DrySupplyID, err)
			}
			if err := s.DrySupplyService.ConsumeStock(ctx, req.DrySupplyID, req.Quantity, prodOrderID); err != nil {
				return fmt.Errorf("consume dry supply %s: %w", req.DrySupplyID, err)
			}
		}
	}
	return nil
}

// reserveAndReady reserves dressed stock for each item and sets sale order to READY_TO_DISPATCH.
func (s *Service) reserveAndReady(ctx context.Context, saleOrderID string, itemStocks []itemStockInfo) error {
	for _, is := range itemStocks {
		inv, err := s.ProductInventoryService.GetProductInventory(is.productID)
		if err != nil {
			return fmt.Errorf("reserveAndReady: get inventory %s: %w", is.productID, err)
		}
		if err := inv.Reserve(saleOrderID, is.requested, userIDFromCtx(ctx)); err != nil {
			return fmt.Errorf("reserveAndReady: reserve stock for %s: %w", is.productID, err)
		}
		if err := s.ProductInventoryService.SaveProductInventory(*inv); err != nil {
			return fmt.Errorf("reserveAndReady: save inventory %s: %w", is.productID, err)
		}
	}
	if _, err := s.SaleOrderService.UpdateSaleOrderStatus(ctx, saleOrderID, sale_order.StatusReadyToDispatch); err != nil {
		return fmt.Errorf("reserveAndReady: update sale order status: %w", err)
	}
	return nil
}

// cancelSaleOrder sets the sale order to CANCELLED.
func (s *Service) cancelSaleOrder(ctx context.Context, saleOrderID string) error {
	_, err := s.SaleOrderService.UpdateSaleOrderStatus(ctx, saleOrderID, sale_order.StatusCancelled)
	return err
}

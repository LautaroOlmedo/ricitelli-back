package application_service

import (
	"context"
	"fmt"
	"testing"

	product_inventory "ricitelli-back/internal/domain/product-inventory"
	sale_order "ricitelli-back/internal/domain/sale-order"
	valueObject "ricitelli-back/internal/value-object"
)

func TestTransitionSaleOrderStatusInventoryBehavior(t *testing.T) {
	tests := []struct {
		name               string
		initialStatus      sale_order.Status
		newStatus          sale_order.Status
		wantPhysical       int64
		wantCommitted      int64
		wantAvailable      int64
		wantInventoryGets  int
		wantInventorySaves int
	}{
		{
			name:               "invoicing preserves physical stock and reservation",
			initialStatus:      sale_order.StatusConfirmed,
			newStatus:          sale_order.StatusInvoiced,
			wantPhysical:       10,
			wantCommitted:      4,
			wantAvailable:      6,
			wantInventoryGets:  0,
			wantInventorySaves: 0,
		},
		{
			name:               "dispatching consumes physical stock and reservation",
			initialStatus:      sale_order.StatusInvoiced,
			newStatus:          sale_order.StatusDispatched,
			wantPhysical:       6,
			wantCommitted:      0,
			wantAvailable:      6,
			wantInventoryGets:  1,
			wantInventorySaves: 1,
		},
		{
			name:               "cancelling releases reservation without consuming physical stock",
			initialStatus:      sale_order.StatusConfirmed,
			newStatus:          sale_order.StatusCancelled,
			wantPhysical:       10,
			wantCommitted:      0,
			wantAvailable:      10,
			wantInventoryGets:  1,
			wantInventorySaves: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			const (
				orderID   = "sale-order-1"
				productID = "product-1"
			)

			order := sale_order.ReconstitueSaleOrder(
				orderID,
				"customer-1",
				tt.initialStatus,
				[]valueObject.SaleOrderItem{{ProductID: productID, Quantity: 4, UnitPrice: 10}},
				sale_order.CurrencyARS,
				sale_order.MarketDomestic,
				"AR",
				sale_order.SaleTypeRegular,
				"2026-06-09T00:00:00Z",
				true,
			)
			inventory := reservedInventory(t, productID, orderID, 10, 4)
			saleOrders := &transitionSaleOrderServiceFake{order: &order}
			inventories := &transitionProductInventoryServiceFake{
				byProductID: map[string]*product_inventory.ProductInventory{productID: &inventory},
			}
			service := Service{
				SaleOrderService:        saleOrders,
				ProductInventoryService: inventories,
			}

			updated, err := service.TransitionSaleOrderStatus(context.Background(), orderID, tt.newStatus)
			if err != nil {
				t.Fatalf("TransitionSaleOrderStatus() error = %v", err)
			}
			if updated.GetStatus() != tt.newStatus {
				t.Fatalf("updated status = %s, want %s", updated.GetStatus(), tt.newStatus)
			}
			if len(saleOrders.updatedStatuses) != 1 || saleOrders.updatedStatuses[0] != tt.newStatus {
				t.Fatalf("status updates = %v, want [%s]", saleOrders.updatedStatuses, tt.newStatus)
			}

			gotInventory := inventories.byProductID[productID]
			assertInventoryState(t, gotInventory, tt.wantPhysical, tt.wantCommitted, tt.wantAvailable)
			if inventories.getCalls != tt.wantInventoryGets {
				t.Fatalf("inventory get calls = %d, want %d", inventories.getCalls, tt.wantInventoryGets)
			}
			if inventories.saveCalls != tt.wantInventorySaves {
				t.Fatalf("inventory save calls = %d, want %d", inventories.saveCalls, tt.wantInventorySaves)
			}
		})
	}
}

func reservedInventory(t *testing.T, productID, orderID string, physical, reserved uint64) product_inventory.ProductInventory {
	t.Helper()

	inventory, err := product_inventory.NewProductInventory(productID, "SKU-1")
	if err != nil {
		t.Fatalf("NewProductInventory() error = %v", err)
	}
	if err := inventory.AddUndressed("production-order-1", physical, "test"); err != nil {
		t.Fatalf("AddUndressed() error = %v", err)
	}
	if err := inventory.ConvertSVtoPT("production-order-1", physical, "LOT-1", "test"); err != nil {
		t.Fatalf("ConvertSVtoPT() error = %v", err)
	}
	if err := inventory.Reserve(orderID, reserved, "test"); err != nil {
		t.Fatalf("Reserve() error = %v", err)
	}
	return inventory
}

func assertInventoryState(t *testing.T, inventory *product_inventory.ProductInventory, physical, committed, available int64) {
	t.Helper()

	if got := inventory.PhysicalDressed(); got != physical {
		t.Errorf("PhysicalDressed() = %d, want %d", got, physical)
	}
	if got := inventory.CommittedDressed(); got != committed {
		t.Errorf("CommittedDressed() = %d, want %d", got, committed)
	}
	gotAvailable, err := inventory.AvailableDressed()
	if err != nil {
		t.Fatalf("AvailableDressed() error = %v", err)
	}
	if gotAvailable != available {
		t.Errorf("AvailableDressed() = %d, want %d", gotAvailable, available)
	}
}

type transitionSaleOrderServiceFake struct {
	order           *sale_order.SaleOrder
	updatedStatuses []sale_order.Status
}

func (f *transitionSaleOrderServiceFake) CreateSaleOrder(context.Context, sale_order.NewSaleOrderParams) (sale_order.SaleOrder, error) {
	return sale_order.SaleOrder{}, fmt.Errorf("unexpected CreateSaleOrder call")
}

func (f *transitionSaleOrderServiceFake) GetSaleOrderByID(context.Context, string) (*sale_order.SaleOrder, error) {
	return f.order, nil
}

func (f *transitionSaleOrderServiceFake) GetSaleOrders(context.Context) ([]sale_order.SaleOrder, error) {
	return nil, fmt.Errorf("unexpected GetSaleOrders call")
}

func (f *transitionSaleOrderServiceFake) UpdateSaleOrderStatus(_ context.Context, _ string, status sale_order.Status) (*sale_order.SaleOrder, error) {
	f.updatedStatuses = append(f.updatedStatuses, status)
	if err := f.order.UpdateStatus(status); err != nil {
		return nil, err
	}
	return f.order, nil
}

type transitionProductInventoryServiceFake struct {
	byProductID map[string]*product_inventory.ProductInventory
	getCalls    int
	saveCalls   int
}

func (f *transitionProductInventoryServiceFake) CreateProductInventory(string, string) error {
	return fmt.Errorf("unexpected CreateProductInventory call")
}

func (f *transitionProductInventoryServiceFake) GetProductInventory(productID string) (*product_inventory.ProductInventory, error) {
	f.getCalls++
	inventory, ok := f.byProductID[productID]
	if !ok {
		return nil, fmt.Errorf("inventory %s not found", productID)
	}
	return inventory, nil
}

func (f *transitionProductInventoryServiceFake) SaveProductInventory(inventory product_inventory.ProductInventory) error {
	f.saveCalls++
	f.byProductID[inventory.GetProductID()] = &inventory
	return nil
}

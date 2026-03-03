package repository

import (
	"context"
	"errors"
	"ricitelli-back/internal/domain/product"
	production_order "ricitelli-back/internal/domain/production-order"
	sale_order "ricitelli-back/internal/domain/sale-order"
	"ricitelli-back/internal/entities"
	valueObject "ricitelli-back/internal/value-object"
	"sync"
)

// InMemoryRepository almacena todo en memoria y es seguro para concurrencia
type InMemoryRepository struct {
	saleOrdersMutex       sync.RWMutex
	productionOrdersMutex sync.RWMutex
	productsMutex         sync.RWMutex

	SaleOrders       []sale_order.SaleOrder
	ProductionOrders []production_order.ProductionOrder
	Products         []product.Product
}

func NewInMemoryRepository() *InMemoryRepository {
	repo := &InMemoryRepository{
		SaleOrders:       make([]sale_order.SaleOrder, 0),
		ProductionOrders: make([]production_order.ProductionOrder, 0),
		Products:         make([]product.Product, 0),
	}

	repo.seedProducts()

	return repo
}

// ========================
// SaleOrderService
// ========================

func (r *InMemoryRepository) CreateSaleOrder(
	ctx context.Context,
	customerID string,
	items []valueObject.SaleOrderItem,
) (sale_order.SaleOrder, error) {
	r.saleOrdersMutex.Lock()
	defer r.saleOrdersMutex.Unlock()

	newOrder, err := sale_order.NewSaleOrder(customerID, items)
	if err != nil {
		return sale_order.SaleOrder{}, err
	}

	r.SaleOrders = append(r.SaleOrders, newOrder)
	return newOrder, nil
}

func (r *InMemoryRepository) GetSaleOrderByID(
	ctx context.Context,
	id string,
) (*sale_order.SaleOrder, error) {
	r.saleOrdersMutex.RLock()
	defer r.saleOrdersMutex.RUnlock()

	for i := range r.SaleOrders {
		if r.SaleOrders[i].GetID() == id {
			return &r.SaleOrders[i], nil
		}
	}
	return nil, errors.New("sale order not found")
}

func (r *InMemoryRepository) GetSaleOrders(ctx context.Context) ([]sale_order.SaleOrder, error) {
	r.saleOrdersMutex.RLock()
	defer r.saleOrdersMutex.RUnlock()

	copies := make([]sale_order.SaleOrder, len(r.SaleOrders))
	copy(copies, r.SaleOrders)
	return copies, nil
}

// ========================
// ProductionOrderService
// ========================

func (r *InMemoryRepository) CreateProductionOrder(
	ctx context.Context,
	saleOrderID string,
	items []entities.ProductionItem,
) error {
	r.productionOrdersMutex.Lock()
	defer r.productionOrdersMutex.Unlock()

	newProdOrder := production_order.NewProductionOrder(saleOrderID, items)
	r.ProductionOrders = append(r.ProductionOrders, newProdOrder)
	return nil
}

func (r *InMemoryRepository) GetProductionOrderByID(
	ctx context.Context,
	id string,
) (*production_order.ProductionOrder, error) {
	r.productionOrdersMutex.RLock()
	defer r.productionOrdersMutex.RUnlock()

	for i := range r.ProductionOrders {
		if r.ProductionOrders[i].GetID() == id {
			return &r.ProductionOrders[i], nil
		}
	}
	return nil, errors.New("production order not found")
}

func (r *InMemoryRepository) GetProductionOrders(ctx context.Context) ([]production_order.ProductionOrder, error) {
	r.productionOrdersMutex.RLock()
	defer r.productionOrdersMutex.RUnlock()

	copies := make([]production_order.ProductionOrder, len(r.ProductionOrders))
	copy(copies, r.ProductionOrders)
	return copies, nil
}

// ========================
// ProductService
// ========================

func (r *InMemoryRepository) CreateProduct(
	ctx context.Context,
	name string,
	bods []valueObject.BillOfDrySupply,
) error {
	r.productsMutex.Lock()
	defer r.productsMutex.Unlock()

	newProduct, err := product.NewProduct(name, bods)
	if err != nil {
		return err
	}
	r.Products = append(r.Products, newProduct)
	return nil
}

func (r *InMemoryRepository) GetProductByID(
	ctx context.Context,
	id string,
) (*product.Product, error) {
	r.productsMutex.RLock()
	defer r.productsMutex.RUnlock()

	for i := range r.Products {
		if r.Products[i].GetID() == id {
			return &r.Products[i], nil
		}
	}
	return nil, errors.New("product not found")
}

func (r *InMemoryRepository) GetProducts(ctx context.Context) ([]product.Product, error) {
	r.productsMutex.RLock()
	defer r.productsMutex.RUnlock()

	copies := make([]product.Product, len(r.Products))
	copy(copies, r.Products)
	return copies, nil
}

// ========================
// MOCKS
// ========================

func (r *InMemoryRepository) seedProducts() {
	products := []struct {
		id   string
		name string
		bom  []valueObject.BillOfDrySupply
	}{
		{
			id:   "wine-malbec-001",
			name: "Malbec Clásico",
			bom: []valueObject.BillOfDrySupply{
				{DrySupplyID: "contraetiqueta_malbec", QuantityPerUnit: 1},
				{DrySupplyID: "corcho", QuantityPerUnit: 1},
				{DrySupplyID: "etiqueta_malbec", QuantityPerUnit: 1},
			},
		},
		{
			id:   "wine-cabernet-002",
			name: "Cabernet Premium",
			bom: []valueObject.BillOfDrySupply{
				{DrySupplyID: "contraetiqueta_cabernet", QuantityPerUnit: 1},
				{DrySupplyID: "corcho", QuantityPerUnit: 1},
				{DrySupplyID: "etiqueta_cabernet", QuantityPerUnit: 2},
				{DrySupplyID: "caja_premium", QuantityPerUnit: 1},
			},
		},
		{
			id:   "wine-reserva-003",
			name: "Reserva Especial",
			bom: []valueObject.BillOfDrySupply{
				{DrySupplyID: "contraetiqueta_dorada", QuantityPerUnit: 1},
				{DrySupplyID: "corcho_reserva", QuantityPerUnit: 1},
				{DrySupplyID: "etiqueta_dorada", QuantityPerUnit: 2},
				{DrySupplyID: "caja_madera", QuantityPerUnit: 1},
			},
		},
	}

	for _, p := range products {
		productCreated, err := product.NewProductWithID(p.id, p.name, p.bom)
		if err != nil {
			panic(err) // válido en seed
		}

		r.Products = append(r.Products, productCreated)
	}
}

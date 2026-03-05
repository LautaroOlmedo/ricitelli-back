package repository

import (
	"context"
	"errors"
	"sync"

	dry_supply "ricitelli-back/internal/domain/dry-supply"
	dry_supply_inventory "ricitelli-back/internal/domain/dry-supply-inventory"
	"ricitelli-back/internal/domain/product"
	product_inventory "ricitelli-back/internal/domain/product-inventory"
	production_order "ricitelli-back/internal/domain/production-order"
	sale_order "ricitelli-back/internal/domain/sale-order"
	"ricitelli-back/internal/entities"
	valueObject "ricitelli-back/internal/value-object"
)

type InMemoryRepository struct {
	saleOrdersMutex         sync.RWMutex
	productionOrdersMutex   sync.RWMutex
	productsMutex           sync.RWMutex
	productInventoryMutex   sync.RWMutex
	drySupplyMutex          sync.RWMutex
	drySupplyInventoryMutex sync.RWMutex

	SaleOrders           []sale_order.SaleOrder
	ProductionOrders     []production_order.ProductionOrder
	Products             []product.Product
	ProductInventory     []product_inventory.ProductInventory
	DrySupplies          []dry_supply.DrySupply
	DrySupplyInventories []dry_supply_inventory.DrySupplyInventory
}

func NewInMemoryRepository() *InMemoryRepository {
	repo := &InMemoryRepository{
		SaleOrders:           make([]sale_order.SaleOrder, 0),
		ProductionOrders:     make([]production_order.ProductionOrder, 0),
		Products:             make([]product.Product, 0),
		ProductInventory:     make([]product_inventory.ProductInventory, 0),
		DrySupplies:          make([]dry_supply.DrySupply, 0),
		DrySupplyInventories: make([]dry_supply_inventory.DrySupplyInventory, 0),
	}
	repo.seedDrySupplies()
	repo.seedProducts()
	repo.seedProductInventory()
	return repo
}

// ===== SALE ORDER =====

func (r *InMemoryRepository) CreateSaleOrder(ctx context.Context, params sale_order.NewSaleOrderParams) (sale_order.SaleOrder, error) {
	r.saleOrdersMutex.Lock()
	defer r.saleOrdersMutex.Unlock()
	newOrder, err := sale_order.NewSaleOrder(params)
	if err != nil {
		return sale_order.SaleOrder{}, err
	}
	r.SaleOrders = append(r.SaleOrders, newOrder)
	return newOrder, nil
}

func (r *InMemoryRepository) GetSaleOrderByID(ctx context.Context, id string) (*sale_order.SaleOrder, error) {
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
	cp := make([]sale_order.SaleOrder, len(r.SaleOrders))
	copy(cp, r.SaleOrders)
	return cp, nil
}

func (r *InMemoryRepository) UpdateSaleOrderStatus(ctx context.Context, id string, newStatus sale_order.Status) (*sale_order.SaleOrder, error) {
	r.saleOrdersMutex.Lock()
	defer r.saleOrdersMutex.Unlock()
	for i := range r.SaleOrders {
		if r.SaleOrders[i].GetID() == id {
			if err := r.SaleOrders[i].UpdateStatus(newStatus); err != nil {
				return nil, err
			}
			return &r.SaleOrders[i], nil
		}
	}
	return nil, errors.New("sale order not found")
}

// ===== PRODUCTION ORDER =====

func (r *InMemoryRepository) CreateProductionOrder(ctx context.Context, saleOrderID string, items []entities.ProductionItem) error {
	r.productionOrdersMutex.Lock()
	defer r.productionOrdersMutex.Unlock()
	newProdOrder := production_order.NewProductionOrder(saleOrderID, items)
	r.ProductionOrders = append(r.ProductionOrders, newProdOrder)
	return nil
}

func (r *InMemoryRepository) GetProductionOrderByID(ctx context.Context, id string) (*production_order.ProductionOrder, error) {
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
	cp := make([]production_order.ProductionOrder, len(r.ProductionOrders))
	copy(cp, r.ProductionOrders)
	return cp, nil
}

// ===== PRODUCT =====

func (r *InMemoryRepository) CreateProduct(ctx context.Context, name string, bods []valueObject.BillOfDrySupply) error {
	r.productsMutex.Lock()
	defer r.productsMutex.Unlock()
	newProduct, err := product.NewProduct(name, bods)
	if err != nil {
		return err
	}
	r.Products = append(r.Products, newProduct)
	return nil
}

func (r *InMemoryRepository) GetProductByID(ctx context.Context, id string) (*product.Product, error) {
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
	cp := make([]product.Product, len(r.Products))
	copy(cp, r.Products)
	return cp, nil
}

// ===== PRODUCT INVENTORY =====

func (r *InMemoryRepository) CreateProductInventory(inv product_inventory.ProductInventory) error {
	r.productInventoryMutex.Lock()
	defer r.productInventoryMutex.Unlock()
	r.ProductInventory = append(r.ProductInventory, inv)
	return nil
}

func (r *InMemoryRepository) GetProductInventory(productID string) (*product_inventory.ProductInventory, error) {
	r.productInventoryMutex.RLock()
	defer r.productInventoryMutex.RUnlock()
	for i := range r.ProductInventory {
		if r.ProductInventory[i].GetProductID() == productID {
			return &r.ProductInventory[i], nil
		}
	}
	return nil, errors.New("product inventory not found")
}

func (r *InMemoryRepository) SaveProductInventory(inv product_inventory.ProductInventory) error {
	r.productInventoryMutex.Lock()
	defer r.productInventoryMutex.Unlock()
	for i := range r.ProductInventory {
		if r.ProductInventory[i].GetProductID() == inv.GetProductID() {
			r.ProductInventory[i] = inv
			return nil
		}
	}
	r.ProductInventory = append(r.ProductInventory, inv)
	return nil
}

// ===== DRY SUPPLY =====

func (r *InMemoryRepository) CreateDrySupply(ctx context.Context, code, name string, category dry_supply.Category, unit string) (*dry_supply.DrySupply, error) {
	r.drySupplyMutex.Lock()
	defer r.drySupplyMutex.Unlock()
	ds, err := dry_supply.NewDrySupply(code, name, category, unit)
	if err != nil {
		return nil, err
	}
	r.DrySupplies = append(r.DrySupplies, ds)
	inv, _ := dry_supply_inventory.NewDrySupplyInventory(ds.GetID())
	r.drySupplyInventoryMutex.Lock()
	r.DrySupplyInventories = append(r.DrySupplyInventories, inv)
	r.drySupplyInventoryMutex.Unlock()
	return &ds, nil
}

func (r *InMemoryRepository) GetDrySupplyByID(ctx context.Context, id string) (*dry_supply.DrySupply, error) {
	r.drySupplyMutex.RLock()
	defer r.drySupplyMutex.RUnlock()
	for i := range r.DrySupplies {
		if r.DrySupplies[i].GetID() == id {
			return &r.DrySupplies[i], nil
		}
	}
	return nil, errors.New("dry supply not found")
}

func (r *InMemoryRepository) GetDrySupplyByCode(ctx context.Context, code string) (*dry_supply.DrySupply, error) {
	r.drySupplyMutex.RLock()
	defer r.drySupplyMutex.RUnlock()
	for i := range r.DrySupplies {
		if r.DrySupplies[i].GetCode() == code {
			return &r.DrySupplies[i], nil
		}
	}
	return nil, errors.New("dry supply not found")
}

func (r *InMemoryRepository) GetDrySupplies(ctx context.Context) ([]dry_supply.DrySupply, error) {
	r.drySupplyMutex.RLock()
	defer r.drySupplyMutex.RUnlock()
	cp := make([]dry_supply.DrySupply, len(r.DrySupplies))
	copy(cp, r.DrySupplies)
	return cp, nil
}

func (r *InMemoryRepository) GetDrySupplyInventory(ctx context.Context, drySupplyID string) (*dry_supply_inventory.DrySupplyInventory, error) {
	r.drySupplyInventoryMutex.RLock()
	defer r.drySupplyInventoryMutex.RUnlock()
	for i := range r.DrySupplyInventories {
		if r.DrySupplyInventories[i].GetDrySupplyID() == drySupplyID {
			return &r.DrySupplyInventories[i], nil
		}
	}
	return nil, errors.New("dry supply inventory not found")
}

func (r *InMemoryRepository) SaveDrySupplyInventory(ctx context.Context, inv dry_supply_inventory.DrySupplyInventory) error {
	r.drySupplyInventoryMutex.Lock()
	defer r.drySupplyInventoryMutex.Unlock()
	for i := range r.DrySupplyInventories {
		if r.DrySupplyInventories[i].GetDrySupplyID() == inv.GetDrySupplyID() {
			r.DrySupplyInventories[i] = inv
			return nil
		}
	}
	r.DrySupplyInventories = append(r.DrySupplyInventories, inv)
	return nil
}

// ===== SEED DATA =====

func (r *InMemoryRepository) seedDrySupplies() {
	type seedDS struct {
		id       string
		code     string
		name     string
		category dry_supply.Category
		unit     string
		stock    uint64
	}
	seeds := []seedDS{
		// Hey Malbec! line
		{"ds-heym-box", "HEYM-BOX-6", "Caja x6 Hey Malbec!", dry_supply.CategoryBox, "UNIT", 2000},
		{"ds-heym-box12", "HEYM-BOX-12", "Caja x12 Hey Malbec!", dry_supply.CategoryBox, "UNIT", 1500},
		{"ds-heym-label", "HEYM-LBL", "Etiqueta Hey Malbec!", dry_supply.CategoryLabel, "UNIT", 15000},
		{"ds-heym-contra", "HEYM-CTR", "Contraetiqueta Hey Malbec! (genérica)", dry_supply.CategoryContraetiqueta, "UNIT", 15000},
		{"ds-heym-contra-jp", "HEYM-CTR-JP", "Contraetiqueta Hey Malbec! Japón", dry_supply.CategoryContraetiqueta, "UNIT", 3000},
		{"ds-heym-contra-sk", "HEYM-CTR-SK", "Contraetiqueta Hey Malbec! Skurnik", dry_supply.CategoryContraetiqueta, "UNIT", 2500},
		{"ds-heym-capsule", "HEYM-CAP", "Cápsula Hey Malbec!", dry_supply.CategoryCapsule, "UNIT", 15000},
		// Kung Fu Malbec line
		{"ds-kungm-box", "KUNGM-BOX-6", "Caja x6 Kung Fu Malbec", dry_supply.CategoryBox, "UNIT", 1200},
		{"ds-kungm-label", "KUNGM-LBL", "Etiqueta Kung Fu Malbec", dry_supply.CategoryLabel, "UNIT", 10000},
		{"ds-kungm-contra", "KUNGM-CTR", "Contraetiqueta Kung Fu Malbec", dry_supply.CategoryContraetiqueta, "UNIT", 10000},
		{"ds-kungm-capsule", "KUNGM-CAP", "Cápsula Kung Fu Malbec", dry_supply.CategoryCapsule, "UNIT", 10000},
		// The Party line
		{"ds-party-box", "TDCL-BOX-6", "Caja x6 The Party", dry_supply.CategoryBox, "UNIT", 800},
		{"ds-party-label", "TDCL-LBL", "Etiqueta The Party", dry_supply.CategoryLabel, "UNIT", 8000},
		{"ds-party-contra", "TDCL-CTR", "Contraetiqueta The Party", dry_supply.CategoryContraetiqueta, "UNIT", 8000},
		// Old Vines Patagonia line
		{"ds-ovp-box", "OVP-BOX-6", "Caja x6 Old Vines Patagonia", dry_supply.CategoryBox, "UNIT", 600},
		{"ds-ovp-label", "OVP-LBL", "Etiqueta Old Vines Patagonia", dry_supply.CategoryLabel, "UNIT", 6000},
		{"ds-ovp-contra", "OVP-CTR", "Contraetiqueta Old Vines Patagonia", dry_supply.CategoryContraetiqueta, "UNIT", 6000},
		// Shared supplies
		{"ds-cork-std", "CORK-STD", "Corcho estándar 38mm", dry_supply.CategoryCork, "UNIT", 50000},
		{"ds-cork-prem", "CORK-PREM", "Corcho premium 44mm", dry_supply.CategoryCork, "UNIT", 20000},
		{"ds-cap-gold", "CAP-GOLD", "Cápsula dorada genérica", dry_supply.CategoryCapsule, "UNIT", 30000},
	}

	for _, s := range seeds {
		ds, err := dry_supply.NewDrySupplyWithID(s.id, s.code, s.name, s.category, s.unit)
		if err != nil {
			panic(err)
		}
		r.DrySupplies = append(r.DrySupplies, ds)

		inv, _ := dry_supply_inventory.NewDrySupplyInventory(ds.GetID())
		if s.stock > 0 {
			_ = inv.AddStock(s.stock, "seed-initial-stock")
		}
		r.DrySupplyInventories = append(r.DrySupplyInventories, inv)
	}
}

func (r *InMemoryRepository) seedProducts() {
	products := []struct {
		id   string
		name string
		bods []valueObject.BillOfDrySupply
	}{
		{
			id:   "HEYM",
			name: "Hey Malbec!",
			bods: []valueObject.BillOfDrySupply{
				{DrySupplyID: "ds-heym-label", QuantityPerUnit: 1},
				{DrySupplyID: "ds-heym-contra", QuantityPerUnit: 1},
				{DrySupplyID: "ds-heym-capsule", QuantityPerUnit: 1},
				{DrySupplyID: "ds-cork-std", QuantityPerUnit: 1},
				{DrySupplyID: "ds-heym-box", QuantityPerUnit: 1},
			},
		},
		{
			id:   "KUNGM",
			name: "Kung Fu Malbec",
			bods: []valueObject.BillOfDrySupply{
				{DrySupplyID: "ds-kungm-label", QuantityPerUnit: 1},
				{DrySupplyID: "ds-kungm-contra", QuantityPerUnit: 1},
				{DrySupplyID: "ds-kungm-capsule", QuantityPerUnit: 1},
				{DrySupplyID: "ds-cork-prem", QuantityPerUnit: 1},
				{DrySupplyID: "ds-kungm-box", QuantityPerUnit: 1},
			},
		},
		{
			id:   "TDCL",
			name: "The Party",
			bods: []valueObject.BillOfDrySupply{
				{DrySupplyID: "ds-party-label", QuantityPerUnit: 1},
				{DrySupplyID: "ds-party-contra", QuantityPerUnit: 1},
				{DrySupplyID: "ds-cap-gold", QuantityPerUnit: 1},
				{DrySupplyID: "ds-cork-std", QuantityPerUnit: 1},
				{DrySupplyID: "ds-party-box", QuantityPerUnit: 1},
			},
		},
		{
			id:   "OVP",
			name: "Old Vines From Patagonia",
			bods: []valueObject.BillOfDrySupply{
				{DrySupplyID: "ds-ovp-label", QuantityPerUnit: 1},
				{DrySupplyID: "ds-ovp-contra", QuantityPerUnit: 1},
				{DrySupplyID: "ds-cap-gold", QuantityPerUnit: 1},
				{DrySupplyID: "ds-cork-prem", QuantityPerUnit: 1},
				{DrySupplyID: "ds-ovp-box", QuantityPerUnit: 1},
			},
		},
	}

	for _, p := range products {
		prod, err := product.NewProductWithID(p.id, p.name, p.bods)
		if err != nil {
			panic(err)
		}
		r.Products = append(r.Products, prod)
	}
}

func (r *InMemoryRepository) seedProductInventory() {
	undressedStock := map[string]uint64{"HEYM": 5000, "KUNGM": 2400, "TDCL": 1800, "OVP": 900}
	dressedStock := map[string]uint64{"HEYM": 2400, "KUNGM": 1200, "TDCL": 900, "OVP": 450}

	for _, prod := range r.Products {
		inv, err := product_inventory.NewProductInventory(prod.GetID(), prod.GetID())
		if err != nil {
			panic(err)
		}
		if qty := undressedStock[prod.GetID()]; qty > 0 {
			_ = inv.AddUndressed("seed-initial-production", qty)
		}
		if qty := dressedStock[prod.GetID()]; qty > 0 {
			_ = inv.ConvertSVtoPT("seed-initial-dressing", qty, "")
		}
		r.ProductInventory = append(r.ProductInventory, inv)
	}
}

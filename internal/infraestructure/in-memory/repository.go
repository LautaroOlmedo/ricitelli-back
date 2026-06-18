package repository

import (
	"context"
	"errors"
	"log"
	"strings"
	"sync"
	"time"

	customer_domain "ricitelli-back/internal/domain/customer"
	dry_supply "ricitelli-back/internal/domain/dry-supply"
	dry_supply_inventory "ricitelli-back/internal/domain/dry-supply-inventory"
	"ricitelli-back/internal/domain/product"
	product_inventory "ricitelli-back/internal/domain/product-inventory"
	production_order "ricitelli-back/internal/domain/production-order"
	sale_order "ricitelli-back/internal/domain/sale-order"
	vineyard_domain "ricitelli-back/internal/domain/vineyard"
	"ricitelli-back/internal/entities"
	inventory_svc "ricitelli-back/internal/service/inventory"
	valueObject "ricitelli-back/internal/value-object"
)

type InMemoryRepository struct {
	administrationOnce sync.Once
	administration     *administrativeStore

	customersMutex          sync.RWMutex
	saleOrdersMutex         sync.RWMutex
	productionOrdersMutex   sync.RWMutex
	productsMutex           sync.RWMutex
	productInventoryMutex   sync.RWMutex
	drySupplyMutex          sync.RWMutex
	drySupplyInventoryMutex sync.RWMutex
	vineyardMutex           sync.RWMutex
	productImagesMutex      sync.RWMutex

	Customers            []customer_domain.Customer
	SaleOrders           []sale_order.SaleOrder
	ProductionOrders     []production_order.ProductionOrder
	Products             []product.Product
	ProductInventory     []product_inventory.ProductInventory
	DrySupplies          []dry_supply.DrySupply
	DrySupplyInventories []dry_supply_inventory.DrySupplyInventory
	Plots                []vineyard_domain.Plot
	ProductImages        map[string]string // product_id → image_url
}

func NewInMemoryRepository() *InMemoryRepository {
	return newRepo(false)
}

func NewInMemoryRepositoryMinimal() *InMemoryRepository {
	return newRepo(true)
}

func newRepo(minimal bool) *InMemoryRepository {
	repo := &InMemoryRepository{
		Customers:            make([]customer_domain.Customer, 0),
		SaleOrders:           make([]sale_order.SaleOrder, 0),
		ProductionOrders:     make([]production_order.ProductionOrder, 0),
		Products:             make([]product.Product, 0),
		ProductInventory:     make([]product_inventory.ProductInventory, 0),
		DrySupplies:          make([]dry_supply.DrySupply, 0),
		DrySupplyInventories: make([]dry_supply_inventory.DrySupplyInventory, 0),
		Plots:                make([]vineyard_domain.Plot, 0),
		ProductImages:        make(map[string]string),
	}

	if minimal {
		log.Println("[InMemoryRepository] minimal mode — solo clientes cargados")
		repo.seedCustomers()
		return repo
	}

	dataPath := resolveDataPath()
	if err := repo.SeedFromXLSX(dataPath); err != nil {
		log.Printf("[InMemoryRepository] xlsx seeding failed (%v); falling back to hardcoded seeds\n", err)
		repo.seedDrySupplies()
		repo.seedProducts()
		repo.seedProductInventory()
		repo.seedCustomers()
	}
	return repo
}

// ===== CUSTOMER =====

func (r *InMemoryRepository) CreateCustomer(ctx context.Context, params customer_domain.NewCustomerParams) (customer_domain.Customer, error) {
	r.customersMutex.Lock()
	defer r.customersMutex.Unlock()
	c, err := customer_domain.NewCustomer(params)
	if err != nil {
		return customer_domain.Customer{}, err
	}
	r.Customers = append(r.Customers, c)
	return c, nil
}

func (r *InMemoryRepository) GetCustomerByID(ctx context.Context, id string) (*customer_domain.Customer, error) {
	r.customersMutex.RLock()
	defer r.customersMutex.RUnlock()
	for i := range r.Customers {
		if r.Customers[i].GetID() == id {
			return &r.Customers[i], nil
		}
	}
	return nil, errors.New("customer not found")
}

func (r *InMemoryRepository) GetCustomers(ctx context.Context) ([]customer_domain.Customer, error) {
	r.customersMutex.RLock()
	defer r.customersMutex.RUnlock()
	cp := make([]customer_domain.Customer, len(r.Customers))
	copy(cp, r.Customers)
	return cp, nil
}

func (r *InMemoryRepository) DeactivateCustomer(ctx context.Context, id string) (*customer_domain.Customer, error) {
	r.customersMutex.Lock()
	defer r.customersMutex.Unlock()
	for i := range r.Customers {
		if r.Customers[i].GetID() == id {
			if err := r.Customers[i].Deactivate(); err != nil {
				return nil, err
			}
			return &r.Customers[i], nil
		}
	}
	return nil, errors.New("customer not found")
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

func (r *InMemoryRepository) CreateProductionOrder(ctx context.Context, saleOrderID string, items []entities.ProductionItem) (production_order.ProductionOrder, error) {
	r.productionOrdersMutex.Lock()
	defer r.productionOrdersMutex.Unlock()
	newProdOrder := production_order.NewProductionOrder(saleOrderID, items)
	r.ProductionOrders = append(r.ProductionOrders, newProdOrder)
	return newProdOrder, nil
}

func (r *InMemoryRepository) UpdateProductionOrderStatus(ctx context.Context, id string, newStatus production_order.Status) (*production_order.ProductionOrder, error) {
	r.productionOrdersMutex.Lock()
	defer r.productionOrdersMutex.Unlock()
	for i := range r.ProductionOrders {
		if r.ProductionOrders[i].GetID() == id {
			if err := r.ProductionOrders[i].UpdateStatus(newStatus); err != nil {
				return nil, err
			}
			return &r.ProductionOrders[i], nil
		}
	}
	return nil, errors.New("production order not found")
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

func (r *InMemoryRepository) CreateProduct(ctx context.Context, name string, bods []valueObject.BillOfDrySupply) (*product.Product, error) {
	r.productsMutex.Lock()
	defer r.productsMutex.Unlock()
	newProduct, err := product.NewProduct(name, bods)
	if err != nil {
		return nil, err
	}
	r.Products = append(r.Products, newProduct)
	return &r.Products[len(r.Products)-1], nil
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

func (r *InMemoryRepository) CreateDrySupply(ctx context.Context, code, name string, category dry_supply.Category, unit string, reorderPoint int) (*dry_supply.DrySupply, error) {
	r.drySupplyMutex.Lock()
	defer r.drySupplyMutex.Unlock()
	ds, err := dry_supply.NewDrySupply(code, name, category, unit, reorderPoint)
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
		id           string
		code         string
		name         string
		category     dry_supply.Category
		unit         string
		stock        uint64
		reorderPoint int
	}
	seeds := []seedDS{
		// Hey Malbec! line
		{"ds-heym-box", "HEYM-BOX-6", "Caja x6 Hey Malbec!", dry_supply.CategoryBox, "UNIT", 2000, 500},
		{"ds-heym-box12", "HEYM-BOX-12", "Caja x12 Hey Malbec!", dry_supply.CategoryBox, "UNIT", 1500, 300},
		{"ds-heym-label", "HEYM-LBL", "Etiqueta Hey Malbec!", dry_supply.CategoryLabel, "UNIT", 15000, 5000},
		{"ds-heym-contra", "HEYM-CTR", "Contraetiqueta Hey Malbec! (genérica)", dry_supply.CategoryContraetiqueta, "UNIT", 15000, 5000},
		{"ds-heym-contra-jp", "HEYM-CTR-JP", "Contraetiqueta Hey Malbec! Japón", dry_supply.CategoryContraetiqueta, "UNIT", 3000, 1000},
		{"ds-heym-contra-sk", "HEYM-CTR-SK", "Contraetiqueta Hey Malbec! Skurnik", dry_supply.CategoryContraetiqueta, "UNIT", 2500, 800},
		{"ds-heym-capsule", "HEYM-CAP", "Cápsula Hey Malbec!", dry_supply.CategoryCapsule, "UNIT", 15000, 5000},
		// Kung Fu Malbec line
		{"ds-kungm-box", "KUNGM-BOX-6", "Caja x6 Kung Fu Malbec", dry_supply.CategoryBox, "UNIT", 1200, 400},
		{"ds-kungm-label", "KUNGM-LBL", "Etiqueta Kung Fu Malbec", dry_supply.CategoryLabel, "UNIT", 10000, 3000},
		{"ds-kungm-contra", "KUNGM-CTR", "Contraetiqueta Kung Fu Malbec", dry_supply.CategoryContraetiqueta, "UNIT", 10000, 3000},
		{"ds-kungm-capsule", "KUNGM-CAP", "Cápsula Kung Fu Malbec", dry_supply.CategoryCapsule, "UNIT", 10000, 3000},
		// The Party line
		{"ds-party-box", "TDCL-BOX-6", "Caja x6 The Party", dry_supply.CategoryBox, "UNIT", 800, 200},
		{"ds-party-label", "TDCL-LBL", "Etiqueta The Party", dry_supply.CategoryLabel, "UNIT", 8000, 2000},
		{"ds-party-contra", "TDCL-CTR", "Contraetiqueta The Party", dry_supply.CategoryContraetiqueta, "UNIT", 8000, 2000},
		// Old Vines Patagonia line
		{"ds-ovp-box", "OVP-BOX-6", "Caja x6 Old Vines Patagonia", dry_supply.CategoryBox, "UNIT", 600, 200},
		{"ds-ovp-label", "OVP-LBL", "Etiqueta Old Vines Patagonia", dry_supply.CategoryLabel, "UNIT", 6000, 1500},
		{"ds-ovp-contra", "OVP-CTR", "Contraetiqueta Old Vines Patagonia", dry_supply.CategoryContraetiqueta, "UNIT", 6000, 1500},
		// Shared supplies
		{"ds-cork-std", "CORK-STD", "Corcho estándar 38mm", dry_supply.CategoryCork, "UNIT", 50000, 10000},
		{"ds-cork-prem", "CORK-PREM", "Corcho premium 44mm", dry_supply.CategoryCork, "UNIT", 20000, 5000},
		{"ds-cap-gold", "CAP-GOLD", "Cápsula dorada genérica", dry_supply.CategoryCapsule, "UNIT", 30000, 8000},
	}

	for _, s := range seeds {
		ds, err := dry_supply.NewDrySupplyWithID(s.id, s.code, s.name, s.category, s.unit, s.reorderPoint)
		if err != nil {
			panic(err)
		}
		r.DrySupplies = append(r.DrySupplies, ds)

		inv, _ := dry_supply_inventory.NewDrySupplyInventory(ds.GetID())
		if s.stock > 0 {
			_ = inv.AddStock(s.stock, "seed-initial-stock", "system")
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

func (r *InMemoryRepository) seedCustomers() {
	seeds := []struct {
		id           string
		socialReason string
		marketType   customer_domain.MarketType
		group        customer_domain.Group
	}{
		{"cust-001", "Distribuidora Sur S.A.", customer_domain.MarketTypeInternal, customer_domain.GroupDistributor},
		{"cust-002", "Vinoteca El Barril", customer_domain.MarketTypeInternal, customer_domain.GroupWineShop},
		{"cust-003", "Japan Wine Imports Ltd.", customer_domain.MarketTypeExternal, customer_domain.GroupExportAgent},
		{"cust-004", "Hotel Patagonia & Spa", customer_domain.MarketTypeInternal, customer_domain.GroupHotel},
	}
	for _, s := range seeds {
		c, err := customer_domain.NewCustomerWithID(s.id, customer_domain.NewCustomerParams{
			SocialReason: s.socialReason,
			MarketType:   s.marketType,
			Group:        s.group,
		})
		if err != nil {
			panic(err)
		}
		r.Customers = append(r.Customers, c)
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
			_ = inv.AddUndressed("seed-initial-production", qty, "system")
		}
		if qty := dressedStock[prod.GetID()]; qty > 0 {
			_ = inv.ConvertSVtoPT("seed-initial-dressing", qty, "", "system")
		}
		r.ProductInventory = append(r.ProductInventory, inv)
	}
}

func (r *InMemoryRepository) GetSaleOrdersByDateRange(ctx context.Context, from, to string) ([]sale_order.SaleOrder, error) {
	r.saleOrdersMutex.RLock()
	defer r.saleOrdersMutex.RUnlock()

	const dateFmt = "2006-01-02"
	fromTime, err := time.Parse(dateFmt, from)
	if err != nil {
		return nil, errors.New("invalid from_date format, expected YYYY-MM-DD")
	}
	toTime, err := time.Parse(dateFmt, to)
	if err != nil {
		return nil, errors.New("invalid to_date format, expected YYYY-MM-DD")
	}
	toTime = toTime.Add(24*time.Hour - time.Nanosecond)

	var result []sale_order.SaleOrder
	for _, o := range r.SaleOrders {
		t, parseErr := time.Parse(time.RFC3339, o.GetCreatedAt())
		if parseErr != nil {
			continue
		}
		if !t.Before(fromTime) && !t.After(toTime) {
			result = append(result, o)
		}
	}
	return result, nil
}

func (r *InMemoryRepository) SearchCustomersBySocialReason(ctx context.Context, query string) ([]customer_domain.Customer, error) {
	r.customersMutex.RLock()
	defer r.customersMutex.RUnlock()

	words := strings.Fields(strings.ToLower(query))
	var result []customer_domain.Customer
	for _, c := range r.Customers {
		name := strings.ToLower(c.GetSocialReason())
		match := true
		for _, w := range words {
			if !strings.Contains(name, w) {
				match = false
				break
			}
		}
		if match {
			result = append(result, c)
		}
	}
	return result, nil
}

// ===== UPDATE METHODS =====

func (r *InMemoryRepository) UpdateCustomer(ctx context.Context, id string, params customer_domain.UpdateCustomerParams) (*customer_domain.Customer, error) {
	r.customersMutex.Lock()
	defer r.customersMutex.Unlock()
	for i := range r.Customers {
		if r.Customers[i].GetID() == id {
			if params.SocialReason != "" {
				if err := r.Customers[i].SetSocialReason(params.SocialReason); err != nil {
					return nil, err
				}
			}
			if params.MarketType != "" {
				if err := r.Customers[i].SetMarketType(params.MarketType); err != nil {
					return nil, err
				}
			}
			if params.Group != "" {
				if err := r.Customers[i].SetGroup(params.Group); err != nil {
					return nil, err
				}
			}
			return &r.Customers[i], nil
		}
	}
	return nil, errors.New("customer not found")
}

func (r *InMemoryRepository) UpdateProduct(ctx context.Context, id, name string, bods []valueObject.BillOfDrySupply) error {
	r.productsMutex.Lock()
	defer r.productsMutex.Unlock()
	for i := range r.Products {
		if r.Products[i].GetID() == id {
			updated, err := product.NewProductWithID(id, name, bods)
			if err != nil {
				return err
			}
			r.Products[i] = updated
			return nil
		}
	}
	return errors.New("product not found")
}

func (r *InMemoryRepository) UpdateDrySupply(ctx context.Context, id, name string, reorderPoint int) error {
	r.drySupplyMutex.Lock()
	defer r.drySupplyMutex.Unlock()
	for i := range r.DrySupplies {
		if r.DrySupplies[i].GetID() == id {
			if name != "" {
				if err := r.DrySupplies[i].SetName(name); err != nil {
					return err
				}
			}
			r.DrySupplies[i].SetReorderPoint(reorderPoint)
			return nil
		}
	}
	return errors.New("dry supply not found")
}

func (r *InMemoryRepository) GetProductionOrdersBySaleOrder(ctx context.Context, saleOrderID string) ([]production_order.ProductionOrder, error) {
	r.productionOrdersMutex.RLock()
	defer r.productionOrdersMutex.RUnlock()
	var result []production_order.ProductionOrder
	for _, o := range r.ProductionOrders {
		if o.GetSalesOrderID() == saleOrderID {
			result = append(result, o)
		}
	}
	return result, nil
}

func (r *InMemoryRepository) GetDailyLotCount(ctx context.Context) (int, error) {
	r.productInventoryMutex.RLock()
	defer r.productInventoryMutex.RUnlock()
	today := time.Now().UTC().Format("2006-01-02")
	count := 0
	for _, inv := range r.ProductInventory {
		for _, m := range inv.GetMovements() {
			if len(m.LotNumber) >= 8 && m.LotNumber[:2] == "L-" {
				// Lot format: L-DDMMYY-NNN-XX → date part is chars 2-7 (DDMMYY)
				// Convert DDMMYY to YYYY-MM-DD for comparison
				if len(m.LotNumber) >= 9 {
					d := m.LotNumber[2:8]
					if len(d) == 6 {
						lotDate := "20" + d[4:6] + "-" + d[2:4] + "-" + d[0:2]
						if lotDate == today {
							count++
						}
					}
				}
			}
		}
	}
	return count, nil
}

// ===== VINEYARD =====

func (r *InMemoryRepository) CreatePlot(ctx context.Context, name, variety string, ha float64, age int32, status string, polygon []vineyard_domain.LatLng) (*vineyard_domain.Plot, error) {
	r.vineyardMutex.Lock()
	defer r.vineyardMutex.Unlock()
	p, err := vineyard_domain.NewPlot(name, variety, ha, age, status, polygon)
	if err != nil {
		return nil, err
	}
	r.Plots = append(r.Plots, p)
	return &r.Plots[len(r.Plots)-1], nil
}

func (r *InMemoryRepository) UpdatePlot(ctx context.Context, id, name, variety string, ha float64, age int32, status string, polygon []vineyard_domain.LatLng) (*vineyard_domain.Plot, error) {
	r.vineyardMutex.Lock()
	defer r.vineyardMutex.Unlock()
	for i := range r.Plots {
		if r.Plots[i].GetID() == id {
			if err := r.Plots[i].Update(name, variety, ha, age, status, polygon); err != nil {
				return nil, err
			}
			return &r.Plots[i], nil
		}
	}
	return nil, vineyard_domain.ErrPlotNotFound
}

func (r *InMemoryRepository) DeletePlot(ctx context.Context, id string) error {
	r.vineyardMutex.Lock()
	defer r.vineyardMutex.Unlock()
	for i := range r.Plots {
		if r.Plots[i].GetID() == id {
			r.Plots = append(r.Plots[:i], r.Plots[i+1:]...)
			return nil
		}
	}
	return vineyard_domain.ErrPlotNotFound
}

func (r *InMemoryRepository) GetPlotByID(ctx context.Context, id string) (*vineyard_domain.Plot, error) {
	r.vineyardMutex.RLock()
	defer r.vineyardMutex.RUnlock()
	for i := range r.Plots {
		if r.Plots[i].GetID() == id {
			return &r.Plots[i], nil
		}
	}
	return nil, vineyard_domain.ErrPlotNotFound
}

func (r *InMemoryRepository) GetPlots(ctx context.Context) ([]vineyard_domain.Plot, error) {
	r.vineyardMutex.RLock()
	defer r.vineyardMutex.RUnlock()
	cp := make([]vineyard_domain.Plot, len(r.Plots))
	copy(cp, r.Plots)
	return cp, nil
}

// ===== PRODUCT IMAGES =====

func (r *InMemoryRepository) SetProductImage(ctx context.Context, id, imageURL string) error {
	r.productImagesMutex.Lock()
	defer r.productImagesMutex.Unlock()
	r.ProductImages[id] = imageURL
	return nil
}

func (r *InMemoryRepository) GetProductImage(ctx context.Context, id string) (string, error) {
	r.productImagesMutex.RLock()
	defer r.productImagesMutex.RUnlock()
	return r.ProductImages[id], nil
}

// QueryMovements returns movements from in-memory data (basic implementation).
func (r *InMemoryRepository) QueryMovements(ctx context.Context, filter inventory_svc.MovementFilter) (*inventory_svc.MovementsResult, error) {
	var entries []inventory_svc.MovementEntry

	includeProducts := filter.Category == "" || filter.Category == "PRODUCT"
	includeSupplies := filter.Category == "" || filter.Category == "DRY_SUPPLY"

	if includeProducts {
		// Build product name lookup
		r.productsMutex.RLock()
		productNames := make(map[string]string)
		for _, p := range r.Products {
			productNames[p.GetID()] = p.GetName()
		}
		r.productsMutex.RUnlock()

		r.productInventoryMutex.RLock()
		for _, inv := range r.ProductInventory {
			for _, m := range inv.GetMovements() {
				if filter.MovementType != "" && string(m.MovementType) != filter.MovementType {
					continue
				}
				if filter.UserID != "" && m.UserID != filter.UserID {
					continue
				}
				if filter.FromDate != "" && m.CreatedAt < filter.FromDate {
					continue
				}
				if filter.ToDate != "" && m.CreatedAt >= filter.ToDate {
					continue
				}
				name := productNames[inv.GetProductID()]
				if name == "" {
					name = inv.GetProductID()
				}
				entries = append(entries, inventory_svc.MovementEntry{
					MovementType: string(m.MovementType),
					Quantity:     m.Quantity,
					Reference:    m.Reference,
					Stage:        string(m.Stage),
					LotNumber:    m.LotNumber,
					UserID:       m.UserID,
					CreatedAt:    m.CreatedAt,
					ItemName:     name,
					Category:     "PRODUCT",
				})
			}
		}
		r.productInventoryMutex.RUnlock()
	}

	if includeSupplies {
		// Build dry supply name lookup
		r.drySupplyMutex.RLock()
		supplyNames := make(map[string]string)
		for _, ds := range r.DrySupplies {
			supplyNames[ds.GetID()] = ds.GetName()
		}
		r.drySupplyMutex.RUnlock()

		r.drySupplyInventoryMutex.RLock()
		for _, inv := range r.DrySupplyInventories {
			for _, m := range inv.GetMovements() {
				if filter.MovementType != "" && string(m.MovementType) != filter.MovementType {
					continue
				}
				if filter.UserID != "" && m.UserID != filter.UserID {
					continue
				}
				if filter.FromDate != "" && m.CreatedAt < filter.FromDate {
					continue
				}
				if filter.ToDate != "" && m.CreatedAt >= filter.ToDate {
					continue
				}
				name := supplyNames[inv.GetDrySupplyID()]
				if name == "" {
					name = inv.GetDrySupplyID()
				}
				entries = append(entries, inventory_svc.MovementEntry{
					MovementType: string(m.MovementType),
					Quantity:     m.Quantity,
					Reference:    m.Reference,
					UserID:       m.UserID,
					CreatedAt:    m.CreatedAt,
					ItemName:     name,
					Category:     "DRY_SUPPLY",
				})
			}
		}
		r.drySupplyInventoryMutex.RUnlock()
	}

	total := len(entries)
	start := (filter.Page - 1) * filter.PageSize
	if start > total {
		start = total
	}
	end := start + filter.PageSize
	if end > total {
		end = total
	}
	return &inventory_svc.MovementsResult{Movements: entries[start:end], TotalCount: total}, nil
}

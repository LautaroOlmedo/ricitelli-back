package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	customer_domain "ricitelli-back/internal/domain/customer"
	dry_supply_domain "ricitelli-back/internal/domain/dry-supply"
	dry_supply_inventory_domain "ricitelli-back/internal/domain/dry-supply-inventory"
	product_domain "ricitelli-back/internal/domain/product"
	product_inventory_domain "ricitelli-back/internal/domain/product-inventory"
	production_order_domain "ricitelli-back/internal/domain/production-order"
	sale_order_domain "ricitelli-back/internal/domain/sale-order"
	"ricitelli-back/internal/entities"
	inventory_svc "ricitelli-back/internal/service/inventory"
	valueObject "ricitelli-back/internal/value-object"
)

// Repository implements all storage interfaces using PostgreSQL.
type Repository struct {
	pool *pgxpool.Pool
}

// NewRepository creates a new PostgreSQL repository and runs migrations.
func NewRepository(databaseURL string) (*Repository, error) {
	cfg, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("parse db config: %w", err)
	}
	cfg.MaxConns = 10
	cfg.MinConns = 2

	pool, err := pgxpool.NewWithConfig(context.Background(), cfg)
	if err != nil {
		return nil, fmt.Errorf("connect to db: %w", err)
	}
	if err := pool.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("ping db: %w", err)
	}
	if err := RunMigrations(pool); err != nil {
		return nil, fmt.Errorf("run migrations: %w", err)
	}
	return &Repository{pool: pool}, nil
}

// ===== CUSTOMER =====

func (r *Repository) CreateCustomer(ctx context.Context, params customer_domain.NewCustomerParams) (customer_domain.Customer, error) {
	c, err := customer_domain.NewCustomer(params)
	if err != nil {
		return customer_domain.Customer{}, err
	}
	_, err = r.pool.Exec(ctx,
		`INSERT INTO customers (id, social_reason, market_type, customer_group, active, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6)`,
		c.GetID(), c.GetSocialReason(), string(c.GetMarketType()), string(c.GetGroup()), c.IsActive(), c.GetCreatedAt(),
	)
	if err != nil {
		return customer_domain.Customer{}, err
	}
	return c, nil
}

func (r *Repository) GetCustomerByID(ctx context.Context, id string) (*customer_domain.Customer, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT id, social_reason, market_type, customer_group, active, created_at FROM customers WHERE id = $1`, id)
	return scanCustomer(row)
}

func (r *Repository) GetCustomers(ctx context.Context) ([]customer_domain.Customer, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, social_reason, market_type, customer_group, active, created_at FROM customers ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return collectCustomers(rows)
}

func (r *Repository) DeactivateCustomer(ctx context.Context, id string) (*customer_domain.Customer, error) {
	row := r.pool.QueryRow(ctx,
		`UPDATE customers SET active = false WHERE id = $1
		 RETURNING id, social_reason, market_type, customer_group, active, created_at`, id)
	return scanCustomer(row)
}

func (r *Repository) UpdateCustomer(ctx context.Context, id string, params customer_domain.UpdateCustomerParams) (*customer_domain.Customer, error) {
	sets := []string{}
	args := []any{}
	i := 1
	if params.SocialReason != "" {
		sets = append(sets, fmt.Sprintf("social_reason = $%d", i))
		args = append(args, params.SocialReason)
		i++
	}
	if params.MarketType != "" {
		sets = append(sets, fmt.Sprintf("market_type = $%d", i))
		args = append(args, string(params.MarketType))
		i++
	}
	if params.Group != "" {
		sets = append(sets, fmt.Sprintf("customer_group = $%d", i))
		args = append(args, string(params.Group))
		i++
	}
	if len(sets) == 0 {
		return r.GetCustomerByID(ctx, id)
	}
	args = append(args, id)
	query := fmt.Sprintf(`UPDATE customers SET %s WHERE id = $%d
		RETURNING id, social_reason, market_type, customer_group, active, created_at`,
		strings.Join(sets, ", "), i)
	row := r.pool.QueryRow(ctx, query, args...)
	return scanCustomer(row)
}

func (r *Repository) SearchCustomersBySocialReason(ctx context.Context, query string) ([]customer_domain.Customer, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, social_reason, market_type, customer_group, active, created_at
		 FROM customers WHERE LOWER(social_reason) LIKE LOWER($1) AND active = true ORDER BY social_reason`,
		"%"+query+"%")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return collectCustomers(rows)
}

func scanCustomer(row pgx.Row) (*customer_domain.Customer, error) {
	var id, socialReason, marketType, group, createdAt string
	var active bool
	if err := row.Scan(&id, &socialReason, &marketType, &group, &active, &createdAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("customer not found")
		}
		return nil, err
	}
	c := customer_domain.ReconstitueCustomer(id, socialReason, customer_domain.MarketType(marketType), customer_domain.Group(group), active, createdAt)
	return &c, nil
}

func collectCustomers(rows pgx.Rows) ([]customer_domain.Customer, error) {
	var result []customer_domain.Customer
	for rows.Next() {
		var id, socialReason, marketType, group, createdAt string
		var active bool
		if err := rows.Scan(&id, &socialReason, &marketType, &group, &active, &createdAt); err != nil {
			return nil, err
		}
		result = append(result, customer_domain.ReconstitueCustomer(id, socialReason, customer_domain.MarketType(marketType), customer_domain.Group(group), active, createdAt))
	}
	return result, rows.Err()
}

// ===== SALE ORDER =====

func (r *Repository) CreateSaleOrder(ctx context.Context, params sale_order_domain.NewSaleOrderParams) (sale_order_domain.SaleOrder, error) {
	so, err := sale_order_domain.NewSaleOrder(params)
	if err != nil {
		return sale_order_domain.SaleOrder{}, err
	}
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return sale_order_domain.SaleOrder{}, err
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx,
		`INSERT INTO sale_orders (id, customer_id, status, currency, market, destination_country, sale_type, active, created_at)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`,
		so.GetID(), so.GetCustomerID(), string(so.GetStatus()), string(so.GetCurrency()),
		string(so.GetMarket()), so.GetDestinationCountry(), string(so.GetSaleType()), so.IsActive(), so.GetCreatedAt(),
	)
	if err != nil {
		return sale_order_domain.SaleOrder{}, err
	}
	for _, item := range so.GetItems() {
		_, err = tx.Exec(ctx,
			`INSERT INTO sale_order_items (sale_order_id, product_id, quantity, unit_price) VALUES ($1,$2,$3,$4)`,
			so.GetID(), item.ProductID, item.Quantity, item.UnitPrice,
		)
		if err != nil {
			return sale_order_domain.SaleOrder{}, err
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return sale_order_domain.SaleOrder{}, err
	}
	return so, nil
}

func (r *Repository) GetSaleOrderByID(ctx context.Context, id string) (*sale_order_domain.SaleOrder, error) {
	return r.loadSaleOrder(ctx, id)
}

func (r *Repository) GetSaleOrders(ctx context.Context) ([]sale_order_domain.SaleOrder, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id FROM sale_orders WHERE active = true ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return r.loadSaleOrders(ctx, ids)
}

func (r *Repository) UpdateSaleOrderStatus(ctx context.Context, id string, newStatus sale_order_domain.Status) (*sale_order_domain.SaleOrder, error) {
	so, err := r.loadSaleOrder(ctx, id)
	if err != nil {
		return nil, err
	}
	if err := so.UpdateStatus(newStatus); err != nil {
		return nil, err
	}
	_, err = r.pool.Exec(ctx, `UPDATE sale_orders SET status = $1 WHERE id = $2`, string(newStatus), id)
	if err != nil {
		return nil, err
	}
	return so, nil
}

func (r *Repository) GetSaleOrdersByDateRange(ctx context.Context, from, to string) ([]sale_order_domain.SaleOrder, error) {
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

	rows, err := r.pool.Query(ctx,
		`SELECT id FROM sale_orders WHERE active = true AND created_at >= $1 AND created_at <= $2 ORDER BY created_at DESC`,
		fromTime.UTC().Format(time.RFC3339), toTime.UTC().Format(time.RFC3339))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return r.loadSaleOrders(ctx, ids)
}

func (r *Repository) loadSaleOrder(ctx context.Context, id string) (*sale_order_domain.SaleOrder, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT id, customer_id, status, currency, market, destination_country, sale_type, created_at, active
		 FROM sale_orders WHERE id = $1`, id)
	var soID, customerID, status, currency, market, destCountry, saleType, createdAt string
	var active bool
	if err := row.Scan(&soID, &customerID, &status, &currency, &market, &destCountry, &saleType, &createdAt, &active); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("sale order not found")
		}
		return nil, err
	}
	items, err := r.loadSaleOrderItems(ctx, soID)
	if err != nil {
		return nil, err
	}
	so := sale_order_domain.ReconstitueSaleOrder(soID, customerID,
		sale_order_domain.Status(status), items,
		sale_order_domain.Currency(currency), sale_order_domain.Market(market),
		destCountry, sale_order_domain.SaleType(saleType), createdAt, active)
	return &so, nil
}

func (r *Repository) loadSaleOrders(ctx context.Context, ids []string) ([]sale_order_domain.SaleOrder, error) {
	result := make([]sale_order_domain.SaleOrder, 0, len(ids))
	for _, id := range ids {
		so, err := r.loadSaleOrder(ctx, id)
		if err != nil {
			return nil, err
		}
		result = append(result, *so)
	}
	return result, nil
}

func (r *Repository) loadSaleOrderItems(ctx context.Context, saleOrderID string) ([]valueObject.SaleOrderItem, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT product_id, quantity, unit_price FROM sale_order_items WHERE sale_order_id = $1`, saleOrderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []valueObject.SaleOrderItem
	for rows.Next() {
		var productID string
		var quantity uint64
		var unitPrice float32
		if err := rows.Scan(&productID, &quantity, &unitPrice); err != nil {
			return nil, err
		}
		items = append(items, valueObject.SaleOrderItem{ProductID: productID, Quantity: quantity, UnitPrice: unitPrice})
	}
	return items, rows.Err()
}

// ===== PRODUCTION ORDER =====

func (r *Repository) CreateProductionOrder(ctx context.Context, salesOrderID string, items []entities.ProductionItem) (production_order_domain.ProductionOrder, error) {
	po := production_order_domain.NewProductionOrder(salesOrderID, items)
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return production_order_domain.ProductionOrder{}, err
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx,
		`INSERT INTO production_orders (id, sale_order_id, operation_number, status, active, created_at)
		 VALUES ($1,$2,$3,$4,$5,$6)`,
		po.GetID(), salesOrderID, po.GetOperationNumber(), string(po.GetStatus()), true, po.GetCreatedAt(),
	)
	if err != nil {
		return production_order_domain.ProductionOrder{}, err
	}
	for _, item := range items {
		var itemID string
		err = tx.QueryRow(ctx,
			`INSERT INTO production_order_items (production_order_id, product_id, quantity) VALUES ($1,$2,$3) RETURNING id`,
			po.GetID(), item.ProductID, item.Quantity,
		).Scan(&itemID)
		if err != nil {
			return production_order_domain.ProductionOrder{}, err
		}
		for _, req := range item.Requirements {
			_, err = tx.Exec(ctx,
				`INSERT INTO production_item_material_requirements (production_order_item_id, dry_supply_id, quantity) VALUES ($1,$2,$3)`,
				itemID, req.DrySupplyID, req.Quantity,
			)
			if err != nil {
				return production_order_domain.ProductionOrder{}, err
			}
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return production_order_domain.ProductionOrder{}, err
	}
	return po, nil
}

func (r *Repository) GetProductionOrderByID(ctx context.Context, id string) (*production_order_domain.ProductionOrder, error) {
	return r.loadProductionOrder(ctx, id)
}

func (r *Repository) GetProductionOrders(ctx context.Context) ([]production_order_domain.ProductionOrder, error) {
	rows, err := r.pool.Query(ctx, `SELECT id FROM production_orders WHERE active = true ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	var result []production_order_domain.ProductionOrder
	for _, id := range ids {
		po, err := r.loadProductionOrder(ctx, id)
		if err != nil {
			return nil, err
		}
		result = append(result, *po)
	}
	return result, nil
}

func (r *Repository) GetProductionOrdersBySaleOrder(ctx context.Context, saleOrderID string) ([]production_order_domain.ProductionOrder, error) {
	rows, err := r.pool.Query(ctx, `SELECT id FROM production_orders WHERE sale_order_id = $1 ORDER BY created_at DESC`, saleOrderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	var result []production_order_domain.ProductionOrder
	for _, id := range ids {
		po, err := r.loadProductionOrder(ctx, id)
		if err != nil {
			return nil, err
		}
		result = append(result, *po)
	}
	return result, nil
}

func (r *Repository) UpdateProductionOrderStatus(ctx context.Context, id string, newStatus production_order_domain.Status) (*production_order_domain.ProductionOrder, error) {
	po, err := r.loadProductionOrder(ctx, id)
	if err != nil {
		return nil, err
	}
	if err := po.UpdateStatus(newStatus); err != nil {
		return nil, err
	}
	_, err = r.pool.Exec(ctx, `UPDATE production_orders SET status = $1 WHERE id = $2`, string(newStatus), id)
	if err != nil {
		return nil, err
	}
	return po, nil
}

func (r *Repository) loadProductionOrder(ctx context.Context, id string) (*production_order_domain.ProductionOrder, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT id, sale_order_id, operation_number, status, created_at, active FROM production_orders WHERE id = $1`, id)
	var poID, saleOrderID, opNumber, status, createdAt string
	var active bool
	if err := row.Scan(&poID, &saleOrderID, &opNumber, &status, &createdAt, &active); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("production order not found")
		}
		return nil, err
	}
	items, err := r.loadProductionItems(ctx, poID)
	if err != nil {
		return nil, err
	}
	po := production_order_domain.ReconstitueProductionOrder(poID, saleOrderID, opNumber, production_order_domain.Status(status), items, createdAt, active)
	return &po, nil
}

func (r *Repository) loadProductionItems(ctx context.Context, productionOrderID string) ([]entities.ProductionItem, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, product_id, quantity FROM production_order_items WHERE production_order_id = $1`, productionOrderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []entities.ProductionItem
	for rows.Next() {
		var itemID, productID string
		var quantity uint64
		if err := rows.Scan(&itemID, &productID, &quantity); err != nil {
			return nil, err
		}
		reqs, err := r.loadMaterialRequirements(ctx, itemID)
		if err != nil {
			return nil, err
		}
		items = append(items, entities.ProductionItem{ProductID: productID, Quantity: quantity, Requirements: reqs})
	}
	return items, rows.Err()
}

func (r *Repository) loadMaterialRequirements(ctx context.Context, productionItemID string) ([]valueObject.MaterialRequirement, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT dry_supply_id, quantity FROM production_item_material_requirements WHERE production_order_item_id = $1`, productionItemID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var reqs []valueObject.MaterialRequirement
	for rows.Next() {
		var dsID string
		var qty uint64
		if err := rows.Scan(&dsID, &qty); err != nil {
			return nil, err
		}
		reqs = append(reqs, valueObject.MaterialRequirement{DrySupplyID: dsID, Quantity: qty})
	}
	return reqs, rows.Err()
}

// ===== PRODUCT =====

func (r *Repository) CreateProduct(ctx context.Context, name string, bods []valueObject.BillOfDrySupply) (*product_domain.Product, error) {
	p, err := product_domain.NewProduct(name, bods)
	if err != nil {
		return nil, err
	}
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)
	_, err = tx.Exec(ctx, `INSERT INTO products (id, name, active) VALUES ($1,$2,$3)`, p.GetID(), p.GetName(), p.GetActive())
	if err != nil {
		return nil, err
	}
	for _, bod := range bods {
		_, err = tx.Exec(ctx,
			`INSERT INTO product_bill_of_dry_supply (product_id, dry_supply_id, quantity_per_unit) VALUES ($1,$2,$3)`,
			p.GetID(), bod.DrySupplyID, bod.QuantityPerUnit,
		)
		if err != nil {
			return nil, err
		}
	}
	// Auto-create inventory record
	_, err = tx.Exec(ctx,
		`INSERT INTO product_inventories (product_id, sku) VALUES ($1,$2) ON CONFLICT (product_id, sku) DO NOTHING`,
		p.GetID(), p.GetID(),
	)
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *Repository) GetProductByID(ctx context.Context, id string) (*product_domain.Product, error) {
	return r.loadProduct(ctx, id)
}

func (r *Repository) GetProducts(ctx context.Context) ([]product_domain.Product, error) {
	rows, err := r.pool.Query(ctx, `SELECT id FROM products WHERE active = true ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	var result []product_domain.Product
	for _, id := range ids {
		p, err := r.loadProduct(ctx, id)
		if err != nil {
			return nil, err
		}
		result = append(result, *p)
	}
	return result, nil
}

func (r *Repository) UpdateProduct(ctx context.Context, id, name string, bods []valueObject.BillOfDrySupply) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	_, err = tx.Exec(ctx, `UPDATE products SET name = $1 WHERE id = $2`, name, id)
	if err != nil {
		return err
	}
	_, err = tx.Exec(ctx, `DELETE FROM product_bill_of_dry_supply WHERE product_id = $1`, id)
	if err != nil {
		return err
	}
	for _, bod := range bods {
		_, err = tx.Exec(ctx,
			`INSERT INTO product_bill_of_dry_supply (product_id, dry_supply_id, quantity_per_unit) VALUES ($1,$2,$3)`,
			id, bod.DrySupplyID, bod.QuantityPerUnit,
		)
		if err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

func (r *Repository) loadProduct(ctx context.Context, id string) (*product_domain.Product, error) {
	row := r.pool.QueryRow(ctx, `SELECT id, name, active FROM products WHERE id = $1`, id)
	var pID, name string
	var active bool
	if err := row.Scan(&pID, &name, &active); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("product not found")
		}
		return nil, err
	}
	bods, err := r.loadBODS(ctx, pID)
	if err != nil {
		return nil, err
	}
	p := product_domain.ReconstitueProduct(pID, name, bods, active)
	return &p, nil
}

func (r *Repository) loadBODS(ctx context.Context, productID string) ([]valueObject.BillOfDrySupply, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT dry_supply_id, quantity_per_unit FROM product_bill_of_dry_supply WHERE product_id = $1`, productID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var bods []valueObject.BillOfDrySupply
	for rows.Next() {
		var dsID string
		var qty uint64
		if err := rows.Scan(&dsID, &qty); err != nil {
			return nil, err
		}
		bods = append(bods, valueObject.BillOfDrySupply{DrySupplyID: dsID, QuantityPerUnit: qty})
	}
	return bods, rows.Err()
}

// ===== PRODUCT INVENTORY =====

func (r *Repository) CreateProductInventory(inv product_inventory_domain.ProductInventory) error {
	_, err := r.pool.Exec(ctx_bg(),
		`INSERT INTO product_inventories (id, product_id, sku) VALUES ($1,$2,$3) ON CONFLICT (product_id, sku) DO NOTHING`,
		inv.GetID(), inv.GetProductID(), inv.GetSku(),
	)
	return err
}

func (r *Repository) GetProductInventory(productID string) (*product_inventory_domain.ProductInventory, error) {
	ctx := ctx_bg()
	row := r.pool.QueryRow(ctx, `SELECT id, product_id, sku FROM product_inventories WHERE product_id = $1`, productID)
	var id, pID, sku string
	if err := row.Scan(&id, &pID, &sku); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("product inventory not found")
		}
		return nil, err
	}
	movements, err := r.loadProductMovements(ctx, id)
	if err != nil {
		return nil, err
	}
	inv := product_inventory_domain.ReconstitueProductInventory(id, pID, sku, movements)
	return &inv, nil
}

func (r *Repository) SaveProductInventory(inv product_inventory_domain.ProductInventory) error {
	ctx := ctx_bg()
	// Upsert the inventory record
	_, err := r.pool.Exec(ctx,
		`INSERT INTO product_inventories (id, product_id, sku) VALUES ($1,$2,$3) ON CONFLICT (product_id, sku) DO NOTHING`,
		inv.GetID(), inv.GetProductID(), inv.GetSku(),
	)
	if err != nil {
		return err
	}
	// Get existing movement count to detect new ones
	var existing int
	_ = r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM product_inventory_movements WHERE product_inventory_id = $1`,
		inv.GetID(),
	).Scan(&existing)

	movements := inv.GetMovements()
	for i := existing; i < len(movements); i++ {
		m := movements[i]
		_, err = r.pool.Exec(ctx,
			`INSERT INTO product_inventory_movements (product_inventory_id, movement_type, quantity, stage, reference, lot_number, user_id, created_at)
			 VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`,
			inv.GetID(), string(m.MovementType), m.Quantity, string(m.Stage), m.Reference, m.LotNumber, m.UserID, m.CreatedAt,
		)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *Repository) loadProductMovements(ctx context.Context, inventoryID string) ([]valueObject.ProductMovement, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT movement_type, quantity, stage, reference, COALESCE(lot_number,''), COALESCE(user_id,''), created_at
		 FROM product_inventory_movements WHERE product_inventory_id = $1 ORDER BY created_at`, inventoryID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var movements []valueObject.ProductMovement
	for rows.Next() {
		var movType, stage, ref, lot, userID, createdAt string
		var qty uint64
		if err := rows.Scan(&movType, &qty, &stage, &ref, &lot, &userID, &createdAt); err != nil {
			return nil, err
		}
		movements = append(movements, valueObject.ProductMovement{
			MovementType: valueObject.ProductMovementType(movType),
			Quantity:     qty,
			Stage:        valueObject.Stage(stage),
			Reference:    ref,
			LotNumber:    lot,
			UserID:       userID,
			CreatedAt:    createdAt,
		})
	}
	return movements, rows.Err()
}

// ===== DRY SUPPLY =====

func (r *Repository) CreateDrySupply(ctx context.Context, code, name string, category dry_supply_domain.Category, unit string, reorderPoint int) (*dry_supply_domain.DrySupply, error) {
	ds, err := dry_supply_domain.NewDrySupply(code, name, category, unit, reorderPoint)
	if err != nil {
		return nil, err
	}
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)
	_, err = tx.Exec(ctx,
		`INSERT INTO dry_supplies (id, code, name, category, unit, reorder_point) VALUES ($1,$2,$3,$4,$5,$6)`,
		ds.GetID(), ds.GetCode(), ds.GetName(), string(ds.GetCategory()), ds.GetUnit(), ds.GetReorderPoint(),
	)
	if err != nil {
		return nil, err
	}
	// Auto-create inventory
	_, err = tx.Exec(ctx,
		`INSERT INTO dry_supply_inventories (id, dry_supply_id) VALUES (gen_random_uuid(), $1)`, ds.GetID())
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return &ds, nil
}

func (r *Repository) GetDrySupplyByID(ctx context.Context, id string) (*dry_supply_domain.DrySupply, error) {
	return r.loadDrySupply(ctx, `SELECT id, code, name, category, unit, reorder_point FROM dry_supplies WHERE id = $1`, id)
}

func (r *Repository) GetDrySupplyByCode(ctx context.Context, code string) (*dry_supply_domain.DrySupply, error) {
	return r.loadDrySupply(ctx, `SELECT id, code, name, category, unit, reorder_point FROM dry_supplies WHERE code = $1`, code)
}

func (r *Repository) GetDrySupplies(ctx context.Context) ([]dry_supply_domain.DrySupply, error) {
	rows, err := r.pool.Query(ctx, `SELECT id, code, name, category, unit, reorder_point FROM dry_supplies ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []dry_supply_domain.DrySupply
	for rows.Next() {
		var id, code, name, category, unit string
		var rp int
		if err := rows.Scan(&id, &code, &name, &category, &unit, &rp); err != nil {
			return nil, err
		}
		result = append(result, dry_supply_domain.ReconstitueDrySupply(id, code, name, dry_supply_domain.Category(category), unit, rp))
	}
	return result, rows.Err()
}

func (r *Repository) UpdateDrySupply(ctx context.Context, id, name string, reorderPoint int) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE dry_supplies SET name = $1, reorder_point = $2 WHERE id = $3`, name, reorderPoint, id)
	return err
}

func (r *Repository) loadDrySupply(ctx context.Context, query string, arg any) (*dry_supply_domain.DrySupply, error) {
	row := r.pool.QueryRow(ctx, query, arg)
	var id, code, name, category, unit string
	var rp int
	if err := row.Scan(&id, &code, &name, &category, &unit, &rp); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("dry supply not found")
		}
		return nil, err
	}
	ds := dry_supply_domain.ReconstitueDrySupply(id, code, name, dry_supply_domain.Category(category), unit, rp)
	return &ds, nil
}

func (r *Repository) GetDrySupplyInventory(ctx context.Context, drySupplyID string) (*dry_supply_inventory_domain.DrySupplyInventory, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT id FROM dry_supply_inventories WHERE dry_supply_id = $1`, drySupplyID)
	var invID string
	if err := row.Scan(&invID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("dry supply inventory not found")
		}
		return nil, err
	}
	movements, err := r.loadDrySupplyMovements(ctx, invID)
	if err != nil {
		return nil, err
	}
	inv := dry_supply_inventory_domain.ReconstitueDrySupplyInventory(invID, drySupplyID, movements)
	return &inv, nil
}

func (r *Repository) SaveDrySupplyInventory(ctx context.Context, inv dry_supply_inventory_domain.DrySupplyInventory) error {
	var existing int
	_ = r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM dry_supply_inventory_movements WHERE dry_supply_inventory_id = $1`,
		inv.GetID(),
	).Scan(&existing)

	movements := inv.GetMovements()
	for i := existing; i < len(movements); i++ {
		m := movements[i]
		_, err := r.pool.Exec(ctx,
			`INSERT INTO dry_supply_inventory_movements (dry_supply_inventory_id, movement_type, quantity, reference, user_id, created_at)
			 VALUES ($1,$2,$3,$4,$5,$6)`,
			inv.GetID(), string(m.MovementType), m.Quantity, m.Reference, m.UserID, m.CreatedAt,
		)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *Repository) loadDrySupplyMovements(ctx context.Context, inventoryID string) ([]valueObject.DrySupplyMovement, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT movement_type, quantity, COALESCE(reference,''), COALESCE(user_id,''), created_at
		 FROM dry_supply_inventory_movements WHERE dry_supply_inventory_id = $1 ORDER BY created_at`, inventoryID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var movements []valueObject.DrySupplyMovement
	for rows.Next() {
		var movType, ref, userID, createdAt string
		var qty uint64
		if err := rows.Scan(&movType, &qty, &ref, &userID, &createdAt); err != nil {
			return nil, err
		}
		movements = append(movements, valueObject.DrySupplyMovement{
			MovementType: valueObject.DrySupplyMovementType(movType),
			Quantity:     qty,
			Reference:    ref,
			UserID:       userID,
			CreatedAt:    createdAt,
		})
	}
	return movements, rows.Err()
}

// GetDailyLotCount counts how many lot numbers have been generated today.
func (r *Repository) GetDailyLotCount(ctx context.Context) (int, error) {
	today := time.Now().UTC().Format("2006-01-02")
	var count int
	err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM product_inventory_movements
		 WHERE lot_number IS NOT NULL AND lot_number != '' AND DATE(created_at) = $1`, today,
	).Scan(&count)
	return count, err
}

// QueryMovements returns a paginated, filtered list of all inventory movements (products + dry supplies).
func (r *Repository) QueryMovements(ctx context.Context, filter inventory_svc.MovementFilter) (*inventory_svc.MovementsResult, error) {
	args := []interface{}{}
	argIdx := 1

	buildWhere := func(dateCol, userCol, typeCol, itemJoinFilter string, itemFilterVal string) string {
		where := ""
		if filter.FromDate != "" {
			where += fmt.Sprintf(" AND %s >= $%d", dateCol, argIdx)
			args = append(args, filter.FromDate)
			argIdx++
		}
		if filter.ToDate != "" {
			where += fmt.Sprintf(" AND %s < $%d", dateCol, argIdx)
			args = append(args, filter.ToDate)
			argIdx++
		}
		if filter.UserID != "" {
			where += fmt.Sprintf(" AND %s = $%d", userCol, argIdx)
			args = append(args, filter.UserID)
			argIdx++
		}
		if filter.MovementType != "" {
			where += fmt.Sprintf(" AND %s = $%d", typeCol, argIdx)
			args = append(args, filter.MovementType)
			argIdx++
		}
		if itemFilterVal != "" {
			where += fmt.Sprintf(" AND %s = $%d", itemJoinFilter, argIdx)
			args = append(args, itemFilterVal)
			argIdx++
		}
		return where
	}

	includeProducts := filter.Category == "" || filter.Category == "PRODUCT"
	includeSupplies := filter.Category == "" || filter.Category == "DRY_SUPPLY"

	queries := []string{}

	if includeProducts {
		pw := buildWhere("m.created_at", "m.user_id", "m.movement_type", "p.id", filter.ProductID)
		q := fmt.Sprintf(`SELECT m.movement_type, m.quantity, m.reference, COALESCE(m.stage,''), COALESCE(m.lot_number,''), COALESCE(m.user_id,''), m.created_at, p.name, 'PRODUCT'
			FROM product_inventory_movements m
			JOIN product_inventories pi ON pi.id = m.product_inventory_id
			JOIN products p ON p.id = pi.product_id
			WHERE 1=1 %s`, pw)
		queries = append(queries, q)
	}

	if includeSupplies {
		sw := buildWhere("m.created_at", "m.user_id", "m.movement_type", "ds.id", filter.DrySupplyID)
		q := fmt.Sprintf(`SELECT m.movement_type, m.quantity, m.reference, '', '', COALESCE(m.user_id,''), m.created_at, ds.name, 'DRY_SUPPLY'
			FROM dry_supply_inventory_movements m
			JOIN dry_supply_inventories dsi ON dsi.id = m.dry_supply_inventory_id
			JOIN dry_supplies ds ON ds.id = dsi.dry_supply_id
			WHERE 1=1 %s`, sw)
		queries = append(queries, q)
	}

	if len(queries) == 0 {
		return &inventory_svc.MovementsResult{}, nil
	}

	unionQuery := strings.Join(queries, " UNION ALL ")

	// Count total
	countSQL := fmt.Sprintf("SELECT COUNT(*) FROM (%s) sub", unionQuery)
	var total int
	if err := r.pool.QueryRow(ctx, countSQL, args...).Scan(&total); err != nil {
		return nil, fmt.Errorf("count movements: %w", err)
	}

	// Paginated data
	offset := (filter.Page - 1) * filter.PageSize
	dataSQL := fmt.Sprintf("%s ORDER BY created_at DESC LIMIT %d OFFSET %d", unionQuery, filter.PageSize, offset)
	rows, err := r.pool.Query(ctx, dataSQL, args...)
	if err != nil {
		return nil, fmt.Errorf("query movements: %w", err)
	}
	defer rows.Close()

	var entries []inventory_svc.MovementEntry
	for rows.Next() {
		var e inventory_svc.MovementEntry
		if err := rows.Scan(&e.MovementType, &e.Quantity, &e.Reference, &e.Stage, &e.LotNumber, &e.UserID, &e.CreatedAt, &e.ItemName, &e.Category); err != nil {
			return nil, err
		}
		entries = append(entries, e)
	}
	return &inventory_svc.MovementsResult{Movements: entries, TotalCount: total}, rows.Err()
}

// ctx_bg is a helper to use context.Background() in methods with no context parameter.
func ctx_bg() context.Context { return context.Background() }

// ===== PRODUCT IMAGES (in-memory fallback until schema migration) =====

var pgProductImages = struct {
	mu   sync.RWMutex
	data map[string]string
}{data: make(map[string]string)}

func (r *Repository) SetProductImage(ctx context.Context, id, imageURL string) error {
	pgProductImages.mu.Lock()
	defer pgProductImages.mu.Unlock()
	pgProductImages.data[id] = imageURL
	return nil
}

func (r *Repository) GetProductImage(ctx context.Context, id string) (string, error) {
	pgProductImages.mu.RLock()
	defer pgProductImages.mu.RUnlock()
	return pgProductImages.data[id], nil
}

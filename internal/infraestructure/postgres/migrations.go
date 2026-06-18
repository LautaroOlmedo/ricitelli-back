package postgres

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
)

// RunMigrations applies any schema changes not covered by the initial init.sql.
// These are additive-only, idempotent ALTER TABLE statements.
func RunMigrations(pool *pgxpool.Pool) error {
	migrations := []struct {
		name string
		sql  string
	}{
		{
			name: "add reorder_point to dry_supplies",
			sql:  `ALTER TABLE dry_supplies ADD COLUMN IF NOT EXISTS reorder_point INTEGER NOT NULL DEFAULT 0`,
		},
		{
			name: "add cancellation_reason to sale_orders",
			sql:  `ALTER TABLE sale_orders ADD COLUMN IF NOT EXISTS cancellation_reason TEXT`,
		},
		{
			name: "add user_id to product_inventory_movements",
			sql:  `ALTER TABLE product_inventory_movements ADD COLUMN IF NOT EXISTS user_id VARCHAR(255) NOT NULL DEFAULT ''`,
		},
		{
			name: "add user_id to dry_supply_inventory_movements",
			sql:  `ALTER TABLE dry_supply_inventory_movements ADD COLUMN IF NOT EXISTS user_id VARCHAR(255) NOT NULL DEFAULT ''`,
		},
		{
			name: "add index on product_inventory_movements user_id",
			sql:  `CREATE INDEX IF NOT EXISTS idx_pim_user_id ON product_inventory_movements(user_id)`,
		},
		{
			name: "add index on product_inventory_movements created_at",
			sql:  `CREATE INDEX IF NOT EXISTS idx_pim_created_at ON product_inventory_movements(created_at)`,
		},
		{
			name: "add index on dry_supply_inventory_movements user_id",
			sql:  `CREATE INDEX IF NOT EXISTS idx_dsim_user_id ON dry_supply_inventory_movements(user_id)`,
		},
		{
			name: "add index on dry_supply_inventory_movements created_at",
			sql:  `CREATE INDEX IF NOT EXISTS idx_dsim_created_at ON dry_supply_inventory_movements(created_at)`,
		},
		{
			name: "add administrative plan V1A and V1B schema",
			sql:  administrativeSchemaSQL,
		},
		{
			name: "persist optional sales invoice on remittances",
			sql: `ALTER TABLE remittances
				ADD COLUMN IF NOT EXISTS sales_invoice_id UUID REFERENCES sales_invoices(id) ON DELETE NO ACTION;
				CREATE INDEX IF NOT EXISTS idx_remittances_sales_invoice_id ON remittances(sales_invoice_id)`,
		},
	}

	for _, m := range migrations {
		if _, err := pool.Exec(context.Background(), m.sql); err != nil {
			return err
		}
		log.Printf("[migration] applied: %s\n", m.name)
	}
	return nil
}

// administrativeSchemaSQL creates the additive administrative schema on
// databases that predate the corresponding init.sql definitions.
const administrativeSchemaSQL = `
CREATE TABLE IF NOT EXISTS sales_invoices (
    id                UUID          NOT NULL DEFAULT gen_random_uuid() PRIMARY KEY,
    customer_id       UUID          NOT NULL REFERENCES customers(id) ON DELETE NO ACTION,
    sale_order_id     UUID          REFERENCES sale_orders(id) ON DELETE NO ACTION,
    document_type     VARCHAR(30)   NOT NULL,
    point_of_sale     INTEGER       NOT NULL CHECK (point_of_sale > 0),
    document_number   BIGINT        NOT NULL CHECK (document_number > 0),
    issue_date        DATE          NOT NULL,
    due_date          DATE,
    currency          VARCHAR(3)    NOT NULL DEFAULT 'ARS',
    exchange_rate     NUMERIC(18,6) NOT NULL DEFAULT 1 CHECK (exchange_rate > 0),
    subtotal          NUMERIC(18,4) NOT NULL CHECK (subtotal >= 0),
    tax_total         NUMERIC(18,4) NOT NULL CHECK (tax_total >= 0),
    total_amount      NUMERIC(18,4) NOT NULL CHECK (total_amount >= 0),
    status            VARCHAR(30)   NOT NULL DEFAULT 'DRAFT'
                                      CHECK (status IN ('DRAFT', 'ISSUED', 'PARTIALLY_PAID', 'PAID', 'CANCELLED')),
    idempotency_key   VARCHAR(255)  NOT NULL UNIQUE,
    created_at        TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    CHECK (due_date IS NULL OR due_date >= issue_date),
    UNIQUE (document_type, point_of_sale, document_number)
);

CREATE TABLE IF NOT EXISTS sales_invoice_items (
    id                UUID          NOT NULL DEFAULT gen_random_uuid() PRIMARY KEY,
    sales_invoice_id  UUID          NOT NULL REFERENCES sales_invoices(id) ON DELETE CASCADE,
    line_number       INTEGER       NOT NULL CHECK (line_number > 0),
    product_id        UUID          REFERENCES products(id) ON DELETE NO ACTION,
    description       VARCHAR(500)  NOT NULL,
    quantity          NUMERIC(18,4) NOT NULL CHECK (quantity > 0),
    unit_price        NUMERIC(18,4) NOT NULL CHECK (unit_price >= 0),
    tax_rate          NUMERIC(7,4)  NOT NULL DEFAULT 0 CHECK (tax_rate >= 0),
    net_amount        NUMERIC(18,4) NOT NULL CHECK (net_amount >= 0),
    tax_amount        NUMERIC(18,4) NOT NULL CHECK (tax_amount >= 0),
    total_amount      NUMERIC(18,4) NOT NULL CHECK (total_amount >= 0),
    created_at        TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    UNIQUE (sales_invoice_id, line_number)
);

CREATE TABLE IF NOT EXISTS remittances (
    id                UUID         NOT NULL DEFAULT gen_random_uuid() PRIMARY KEY,
    customer_id       UUID         NOT NULL REFERENCES customers(id) ON DELETE NO ACTION,
    sale_order_id     UUID         REFERENCES sale_orders(id) ON DELETE NO ACTION,
    sales_invoice_id  UUID         REFERENCES sales_invoices(id) ON DELETE NO ACTION,
    point_of_sale     INTEGER      NOT NULL CHECK (point_of_sale > 0),
    document_number   BIGINT       NOT NULL CHECK (document_number > 0),
    issue_date        DATE         NOT NULL,
    delivery_date     DATE,
    status            VARCHAR(30)  NOT NULL DEFAULT 'DRAFT'
                                    CHECK (status IN ('DRAFT', 'ISSUED', 'DELIVERED', 'CANCELLED')),
    idempotency_key   VARCHAR(255) NOT NULL UNIQUE,
    created_at        TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    CHECK (delivery_date IS NULL OR delivery_date >= issue_date),
    UNIQUE (point_of_sale, document_number)
);

CREATE TABLE IF NOT EXISTS remittance_items (
    id                UUID          NOT NULL DEFAULT gen_random_uuid() PRIMARY KEY,
    remittance_id     UUID          NOT NULL REFERENCES remittances(id) ON DELETE CASCADE,
    line_number       INTEGER       NOT NULL CHECK (line_number > 0),
    product_id        UUID          REFERENCES products(id) ON DELETE NO ACTION,
    description       VARCHAR(500)  NOT NULL,
    quantity          NUMERIC(18,4) NOT NULL CHECK (quantity > 0),
    lot_number        VARCHAR(100),
    created_at        TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    UNIQUE (remittance_id, line_number)
);

CREATE TABLE IF NOT EXISTS customer_receipts (
    id                UUID          NOT NULL DEFAULT gen_random_uuid() PRIMARY KEY,
    customer_id       UUID          NOT NULL REFERENCES customers(id) ON DELETE NO ACTION,
    receipt_number    BIGINT        NOT NULL CHECK (receipt_number > 0) UNIQUE,
    receipt_date      DATE          NOT NULL,
    currency          VARCHAR(3)    NOT NULL DEFAULT 'ARS',
    exchange_rate     NUMERIC(18,6) NOT NULL DEFAULT 1 CHECK (exchange_rate > 0),
    amount            NUMERIC(18,4) NOT NULL CHECK (amount > 0),
    payment_method    VARCHAR(50)   NOT NULL,
    payment_reference VARCHAR(255),
    status            VARCHAR(30)   NOT NULL DEFAULT 'DRAFT'
                                      CHECK (status IN ('DRAFT', 'POSTED', 'VOIDED')),
    idempotency_key   VARCHAR(255)  NOT NULL UNIQUE,
    created_at        TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ   NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS customer_receipt_allocations (
    id                  UUID          NOT NULL DEFAULT gen_random_uuid() PRIMARY KEY,
    customer_receipt_id UUID          NOT NULL REFERENCES customer_receipts(id) ON DELETE CASCADE,
    sales_invoice_id    UUID          NOT NULL REFERENCES sales_invoices(id) ON DELETE NO ACTION,
    allocated_amount    NUMERIC(18,4) NOT NULL CHECK (allocated_amount > 0),
    created_at          TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    UNIQUE (customer_receipt_id, sales_invoice_id)
);

CREATE TABLE IF NOT EXISTS suppliers (
    id                UUID         NOT NULL DEFAULT gen_random_uuid() PRIMARY KEY,
    tax_id            VARCHAR(50)  NOT NULL UNIQUE,
    social_reason     VARCHAR(255) NOT NULL,
    trade_name        VARCHAR(255),
    email             VARCHAR(255),
    phone             VARCHAR(100),
    address           VARCHAR(500),
    active            BOOLEAN      NOT NULL DEFAULT TRUE,
    idempotency_key   VARCHAR(255) NOT NULL UNIQUE,
    created_at        TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS purchase_needs (
    id                UUID         NOT NULL DEFAULT gen_random_uuid() PRIMARY KEY,
    need_number       BIGINT       NOT NULL CHECK (need_number > 0) UNIQUE,
    requested_date    DATE         NOT NULL,
    required_by_date  DATE,
    status            VARCHAR(30)  NOT NULL DEFAULT 'DRAFT'
                                    CHECK (status IN ('DRAFT', 'OPEN', 'QUOTED', 'ORDERED', 'CLOSED', 'CANCELLED')),
    notes             TEXT,
    idempotency_key   VARCHAR(255) NOT NULL UNIQUE,
    created_at        TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    CHECK (required_by_date IS NULL OR required_by_date >= requested_date)
);

CREATE TABLE IF NOT EXISTS purchase_need_items (
    id                UUID          NOT NULL DEFAULT gen_random_uuid() PRIMARY KEY,
    purchase_need_id  UUID          NOT NULL REFERENCES purchase_needs(id) ON DELETE CASCADE,
    line_number       INTEGER       NOT NULL CHECK (line_number > 0),
    dry_supply_id     UUID          REFERENCES dry_supplies(id) ON DELETE NO ACTION,
    description       VARCHAR(500)  NOT NULL,
    quantity          NUMERIC(18,4) NOT NULL CHECK (quantity > 0),
    unit              VARCHAR(50)   NOT NULL,
    created_at        TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    UNIQUE (purchase_need_id, line_number)
);

CREATE TABLE IF NOT EXISTS supplier_quotes (
    id                UUID          NOT NULL DEFAULT gen_random_uuid() PRIMARY KEY,
    supplier_id       UUID          NOT NULL REFERENCES suppliers(id) ON DELETE NO ACTION,
    purchase_need_id  UUID          REFERENCES purchase_needs(id) ON DELETE NO ACTION,
    quote_number      VARCHAR(100)  NOT NULL,
    quote_date        DATE          NOT NULL,
    valid_until       DATE,
    currency          VARCHAR(3)    NOT NULL DEFAULT 'ARS',
    exchange_rate     NUMERIC(18,6) NOT NULL DEFAULT 1 CHECK (exchange_rate > 0),
    subtotal          NUMERIC(18,4) NOT NULL CHECK (subtotal >= 0),
    tax_total         NUMERIC(18,4) NOT NULL CHECK (tax_total >= 0),
    total_amount      NUMERIC(18,4) NOT NULL CHECK (total_amount >= 0),
    status            VARCHAR(30)   NOT NULL DEFAULT 'DRAFT'
                                      CHECK (status IN ('DRAFT', 'RECEIVED', 'ACCEPTED', 'REJECTED', 'EXPIRED', 'CANCELLED')),
    idempotency_key   VARCHAR(255)  NOT NULL UNIQUE,
    created_at        TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    CHECK (valid_until IS NULL OR valid_until >= quote_date),
    UNIQUE (supplier_id, quote_number)
);

CREATE TABLE IF NOT EXISTS supplier_quote_items (
    id                    UUID          NOT NULL DEFAULT gen_random_uuid() PRIMARY KEY,
    supplier_quote_id     UUID          NOT NULL REFERENCES supplier_quotes(id) ON DELETE CASCADE,
    purchase_need_item_id UUID          REFERENCES purchase_need_items(id) ON DELETE NO ACTION,
    line_number           INTEGER       NOT NULL CHECK (line_number > 0),
    dry_supply_id         UUID          REFERENCES dry_supplies(id) ON DELETE NO ACTION,
    description           VARCHAR(500)  NOT NULL,
    quantity              NUMERIC(18,4) NOT NULL CHECK (quantity > 0),
    unit                  VARCHAR(50)   NOT NULL,
    unit_price            NUMERIC(18,4) NOT NULL CHECK (unit_price >= 0),
    tax_rate              NUMERIC(7,4)  NOT NULL DEFAULT 0 CHECK (tax_rate >= 0),
    net_amount            NUMERIC(18,4) NOT NULL CHECK (net_amount >= 0),
    tax_amount            NUMERIC(18,4) NOT NULL CHECK (tax_amount >= 0),
    total_amount          NUMERIC(18,4) NOT NULL CHECK (total_amount >= 0),
    created_at            TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    updated_at            TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    UNIQUE (supplier_quote_id, line_number)
);

CREATE TABLE IF NOT EXISTS supplier_invoices (
    id                UUID          NOT NULL DEFAULT gen_random_uuid() PRIMARY KEY,
    supplier_id       UUID          NOT NULL REFERENCES suppliers(id) ON DELETE NO ACTION,
    supplier_quote_id UUID          REFERENCES supplier_quotes(id) ON DELETE NO ACTION,
    document_type     VARCHAR(30)   NOT NULL,
    point_of_sale     INTEGER       NOT NULL CHECK (point_of_sale > 0),
    document_number   BIGINT        NOT NULL CHECK (document_number > 0),
    issue_date        DATE          NOT NULL,
    due_date          DATE,
    currency          VARCHAR(3)    NOT NULL DEFAULT 'ARS',
    exchange_rate     NUMERIC(18,6) NOT NULL DEFAULT 1 CHECK (exchange_rate > 0),
    subtotal          NUMERIC(18,4) NOT NULL CHECK (subtotal >= 0),
    tax_total         NUMERIC(18,4) NOT NULL CHECK (tax_total >= 0),
    total_amount      NUMERIC(18,4) NOT NULL CHECK (total_amount >= 0),
    status            VARCHAR(30)   NOT NULL DEFAULT 'DRAFT'
                                      CHECK (status IN ('DRAFT', 'ISSUED', 'PARTIALLY_PAID', 'PAID', 'CANCELLED')),
    idempotency_key   VARCHAR(255)  NOT NULL UNIQUE,
    created_at        TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    CHECK (due_date IS NULL OR due_date >= issue_date),
    UNIQUE (supplier_id, document_type, point_of_sale, document_number)
);

CREATE TABLE IF NOT EXISTS supplier_invoice_items (
    id                     UUID          NOT NULL DEFAULT gen_random_uuid() PRIMARY KEY,
    supplier_invoice_id    UUID          NOT NULL REFERENCES supplier_invoices(id) ON DELETE CASCADE,
    supplier_quote_item_id UUID          REFERENCES supplier_quote_items(id) ON DELETE NO ACTION,
    line_number            INTEGER       NOT NULL CHECK (line_number > 0),
    dry_supply_id          UUID          REFERENCES dry_supplies(id) ON DELETE NO ACTION,
    description            VARCHAR(500)  NOT NULL,
    quantity               NUMERIC(18,4) NOT NULL CHECK (quantity > 0),
    unit                   VARCHAR(50)   NOT NULL,
    unit_price             NUMERIC(18,4) NOT NULL CHECK (unit_price >= 0),
    tax_rate               NUMERIC(7,4)  NOT NULL DEFAULT 0 CHECK (tax_rate >= 0),
    net_amount             NUMERIC(18,4) NOT NULL CHECK (net_amount >= 0),
    tax_amount             NUMERIC(18,4) NOT NULL CHECK (tax_amount >= 0),
    total_amount           NUMERIC(18,4) NOT NULL CHECK (total_amount >= 0),
    created_at             TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    updated_at             TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    UNIQUE (supplier_invoice_id, line_number)
);

CREATE TABLE IF NOT EXISTS supplier_payments (
    id                UUID          NOT NULL DEFAULT gen_random_uuid() PRIMARY KEY,
    supplier_id       UUID          NOT NULL REFERENCES suppliers(id) ON DELETE NO ACTION,
    payment_number    BIGINT        NOT NULL CHECK (payment_number > 0) UNIQUE,
    payment_date      DATE          NOT NULL,
    currency          VARCHAR(3)    NOT NULL DEFAULT 'ARS',
    exchange_rate     NUMERIC(18,6) NOT NULL DEFAULT 1 CHECK (exchange_rate > 0),
    amount            NUMERIC(18,4) NOT NULL CHECK (amount > 0),
    payment_method    VARCHAR(50)   NOT NULL,
    payment_reference VARCHAR(255),
    status            VARCHAR(30)   NOT NULL DEFAULT 'DRAFT'
                                      CHECK (status IN ('DRAFT', 'POSTED', 'VOIDED')),
    idempotency_key   VARCHAR(255)  NOT NULL UNIQUE,
    created_at        TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ   NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS supplier_payment_allocations (
    id                  UUID          NOT NULL DEFAULT gen_random_uuid() PRIMARY KEY,
    supplier_payment_id UUID          NOT NULL REFERENCES supplier_payments(id) ON DELETE CASCADE,
    supplier_invoice_id UUID          NOT NULL REFERENCES supplier_invoices(id) ON DELETE NO ACTION,
    allocated_amount    NUMERIC(18,4) NOT NULL CHECK (allocated_amount > 0),
    created_at          TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    UNIQUE (supplier_payment_id, supplier_invoice_id)
);

CREATE INDEX IF NOT EXISTS idx_sales_invoices_customer_id              ON sales_invoices(customer_id);
CREATE INDEX IF NOT EXISTS idx_sales_invoices_sale_order_id            ON sales_invoices(sale_order_id);
CREATE INDEX IF NOT EXISTS idx_sales_invoices_issue_date               ON sales_invoices(issue_date);
CREATE INDEX IF NOT EXISTS idx_sales_invoices_status                   ON sales_invoices(status);
CREATE INDEX IF NOT EXISTS idx_sales_invoice_items_invoice_id          ON sales_invoice_items(sales_invoice_id);
CREATE INDEX IF NOT EXISTS idx_sales_invoice_items_product_id          ON sales_invoice_items(product_id);
CREATE INDEX IF NOT EXISTS idx_remittances_customer_id                 ON remittances(customer_id);
CREATE INDEX IF NOT EXISTS idx_remittances_sale_order_id               ON remittances(sale_order_id);
CREATE INDEX IF NOT EXISTS idx_remittances_issue_date                  ON remittances(issue_date);
CREATE INDEX IF NOT EXISTS idx_remittance_items_remittance_id          ON remittance_items(remittance_id);
CREATE INDEX IF NOT EXISTS idx_remittance_items_product_id             ON remittance_items(product_id);
CREATE INDEX IF NOT EXISTS idx_customer_receipts_customer_id           ON customer_receipts(customer_id);
CREATE INDEX IF NOT EXISTS idx_customer_receipts_receipt_date          ON customer_receipts(receipt_date);
CREATE INDEX IF NOT EXISTS idx_customer_receipt_allocations_invoice_id ON customer_receipt_allocations(sales_invoice_id);
CREATE INDEX IF NOT EXISTS idx_suppliers_social_reason                 ON suppliers(social_reason);
CREATE INDEX IF NOT EXISTS idx_purchase_needs_status                   ON purchase_needs(status);
CREATE INDEX IF NOT EXISTS idx_purchase_needs_requested_date           ON purchase_needs(requested_date);
CREATE INDEX IF NOT EXISTS idx_purchase_need_items_dry_supply_id       ON purchase_need_items(dry_supply_id);
CREATE INDEX IF NOT EXISTS idx_supplier_quotes_supplier_id             ON supplier_quotes(supplier_id);
CREATE INDEX IF NOT EXISTS idx_supplier_quotes_purchase_need_id        ON supplier_quotes(purchase_need_id);
CREATE INDEX IF NOT EXISTS idx_supplier_quotes_status                  ON supplier_quotes(status);
CREATE INDEX IF NOT EXISTS idx_supplier_quote_items_dry_supply_id      ON supplier_quote_items(dry_supply_id);
CREATE INDEX IF NOT EXISTS idx_supplier_quote_items_need_item_id       ON supplier_quote_items(purchase_need_item_id);
CREATE INDEX IF NOT EXISTS idx_supplier_invoices_supplier_id           ON supplier_invoices(supplier_id);
CREATE INDEX IF NOT EXISTS idx_supplier_invoices_quote_id              ON supplier_invoices(supplier_quote_id);
CREATE INDEX IF NOT EXISTS idx_supplier_invoices_issue_date            ON supplier_invoices(issue_date);
CREATE INDEX IF NOT EXISTS idx_supplier_invoices_status                ON supplier_invoices(status);
CREATE INDEX IF NOT EXISTS idx_supplier_invoice_items_dry_supply_id    ON supplier_invoice_items(dry_supply_id);
CREATE INDEX IF NOT EXISTS idx_supplier_invoice_items_quote_item_id    ON supplier_invoice_items(supplier_quote_item_id);
CREATE INDEX IF NOT EXISTS idx_supplier_payments_supplier_id           ON supplier_payments(supplier_id);
CREATE INDEX IF NOT EXISTS idx_supplier_payments_payment_date          ON supplier_payments(payment_date);
CREATE INDEX IF NOT EXISTS idx_supplier_payment_allocations_invoice_id ON supplier_payment_allocations(supplier_invoice_id);
`

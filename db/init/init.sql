-- =============================================================================
-- RICITELLI BACK - Database Schema
-- PostgreSQL 13+
-- =============================================================================
--
-- DOMAIN MODEL SUMMARY
-- --------------------
-- Aggregates  : customers, dry_supplies, products, sale_orders,
--               production_orders, dry_supply_inventories, product_inventories
-- Entities    : production_order_items
-- Value Objects (denormalized into tables):
--               product_bill_of_dry_supply               (inside Product)
--               sale_order_items                         (inside SaleOrder)
--               production_item_material_requirements    (inside ProductionItem)
--               dry_supply_inventory_movements           (inside DrySupplyInventory)
--               product_inventory_movements              (inside ProductInventory)
--
-- RELATIONSHIP MAP
-- ----------------
-- customers                          1 : N  sale_orders
-- sale_orders                        1 : N  sale_order_items
-- sale_order_items                   N : 1  products
-- sale_orders                        1 : N  production_orders
-- products                           1 : N  product_bill_of_dry_supply
-- product_bill_of_dry_supply         N : 1  dry_supplies
-- dry_supplies                       1 : 1  dry_supply_inventories
-- dry_supply_inventories             1 : N  dry_supply_inventory_movements
-- products                           1 : N  product_inventories
-- product_inventories                1 : N  product_inventory_movements
-- production_orders                  1 : N  production_order_items
-- production_order_items             N : 1  products
-- production_order_items             1 : N  production_item_material_requirements
-- production_item_material_requirements N : 1  dry_supplies
-- =============================================================================

-- NOTE: gen_random_uuid() is built-in since PostgreSQL 13. No extensions required.

-- ---------------------------------------------------------------------------
-- AGGREGATE: customers
-- Root aggregate representing a buyer who places sale orders.
-- Relationship: 1:N with sale_orders
-- ---------------------------------------------------------------------------
CREATE TABLE customers (
    id             UUID         NOT NULL DEFAULT gen_random_uuid() PRIMARY KEY,
    -- razon social del cliente
    social_reason  VARCHAR(255) NOT NULL,
    -- valores: INTERNAL, EXTERNAL
    market_type    VARCHAR(50)  NOT NULL,
    -- valores: DISTRIBUTOR, WINE_SHOP, RESTAURANT, HOTEL, RETAIL, PRIVATE, EXPORT_AGENT
    customer_group VARCHAR(50)  NOT NULL,
    active         BOOLEAN      NOT NULL DEFAULT TRUE,
    -- fecha en formato RFC3339, ej: 2024-01-15T10:30:00Z
    created_at     VARCHAR(50)  NOT NULL
);

-- ---------------------------------------------------------------------------
-- AGGREGATE: dry_supplies
-- Raw/packaging materials: labels, corks, capsules, boxes, bottles, etc.
-- Relationship: 1:1 with dry_supply_inventories
--               1:N with product_bill_of_dry_supply
--               1:N with production_item_material_requirements
-- ---------------------------------------------------------------------------
CREATE TABLE dry_supplies (
    id          UUID         NOT NULL DEFAULT gen_random_uuid() PRIMARY KEY,
    -- SKU/code, ej: IF1156
    code        VARCHAR(100) NOT NULL UNIQUE,
    name        VARCHAR(255) NOT NULL,
    -- valores: LABEL, CONTRAETIQUETA, BOX, CORK, CAPSULE, BOTTLE, OTHER
    category    VARCHAR(50)  NOT NULL,
    description VARCHAR(500),
    -- valores: UNIT, BOX, KG
    unit        VARCHAR(50)  NOT NULL
);

-- ---------------------------------------------------------------------------
-- AGGREGATE: products
-- Manufacturable wine products defined by name and bill of materials.
-- Relationship: 1:N with product_bill_of_dry_supply
--               1:N with product_inventories
--               1:N with sale_order_items
--               1:N with production_order_items
-- ---------------------------------------------------------------------------
CREATE TABLE products (
    id     UUID         NOT NULL DEFAULT gen_random_uuid() PRIMARY KEY,
    name   VARCHAR(255) NOT NULL,
    active BOOLEAN      NOT NULL DEFAULT TRUE
);

-- ---------------------------------------------------------------------------
-- VALUE OBJECT: product_bill_of_dry_supply  (BillOfDrySupply inside Product)
-- Bill of materials: qty of each dry supply needed to produce one unit of product.
-- Relationship: N:1 with products
--               N:1 with dry_supplies
--               (M:N enriched table between products and dry_supplies)
-- ---------------------------------------------------------------------------
CREATE TABLE product_bill_of_dry_supply (
    id                UUID   NOT NULL DEFAULT gen_random_uuid() PRIMARY KEY,
    product_id        UUID   NOT NULL REFERENCES products(id)     ON DELETE CASCADE,
    dry_supply_id     UUID   NOT NULL REFERENCES dry_supplies(id) ON DELETE NO ACTION,
    quantity_per_unit BIGINT NOT NULL CHECK (quantity_per_unit > 0),
    UNIQUE (product_id, dry_supply_id)
);

-- ---------------------------------------------------------------------------
-- AGGREGATE: sale_orders
-- Root aggregate for customer sales orders.
-- Status pipeline: NEW -> CONFIRMED -> INVOICED -> DISPATCHED (or CANCELLED)
-- Relationship: N:1 with customers
--               1:N with sale_order_items
--               1:N with production_orders
-- ---------------------------------------------------------------------------
CREATE TABLE sale_orders (
    id                  UUID        NOT NULL DEFAULT gen_random_uuid() PRIMARY KEY,
    customer_id         UUID        NOT NULL REFERENCES customers(id) ON DELETE NO ACTION,
    -- valores: NEW, CONFIRMED, INVOICED, DISPATCHED, CANCELLED
    status              VARCHAR(50) NOT NULL DEFAULT 'NEW',
    -- valores: ARS, USD, CAD, EUR
    currency            VARCHAR(10) NOT NULL DEFAULT 'ARS',
    -- valores: DOMESTIC, EXPORT
    market              VARCHAR(20) NOT NULL DEFAULT 'DOMESTIC',
    -- ISO-3166 alpha-2, ej: AR, GB, BR, JP
    destination_country VARCHAR(10),
    -- valores: SALE, SAMPLE_CUSTOMS, GIFT, INTERNAL, COMMERCIAL_SAMPLE
    sale_type           VARCHAR(50) NOT NULL DEFAULT 'SALE',
    active              BOOLEAN     NOT NULL DEFAULT TRUE,
    -- fecha en formato RFC3339
    created_at          VARCHAR(50) NOT NULL
);

-- ---------------------------------------------------------------------------
-- VALUE OBJECT: sale_order_items  (SaleOrderItem inside SaleOrder)
-- Individual line items within a sale order.
-- Relationship: N:1 with sale_orders
--               N:1 with products
-- ---------------------------------------------------------------------------
CREATE TABLE sale_order_items (
    id            UUID          NOT NULL DEFAULT gen_random_uuid() PRIMARY KEY,
    sale_order_id UUID          NOT NULL REFERENCES sale_orders(id) ON DELETE CASCADE,
    product_id    UUID          NOT NULL REFERENCES products(id)    ON DELETE NO ACTION,
    quantity      BIGINT        NOT NULL CHECK (quantity > 0),
    unit_price    NUMERIC(12,4) NOT NULL CHECK (unit_price >= 0)
);

-- ---------------------------------------------------------------------------
-- AGGREGATE: production_orders
-- Manufacturing order derived from a sale order.
-- Relationship: N:1 with sale_orders
--               1:N with production_order_items
-- ---------------------------------------------------------------------------
CREATE TABLE production_orders (
    id               UUID         NOT NULL DEFAULT gen_random_uuid() PRIMARY KEY,
    sale_order_id    UUID         NOT NULL REFERENCES sale_orders(id) ON DELETE NO ACTION,
    operation_number VARCHAR(100) NOT NULL,
    status           VARCHAR(50)  NOT NULL DEFAULT 'IN PROGRESS',
    active           BOOLEAN      NOT NULL DEFAULT TRUE,
    -- fecha en formato RFC3339
    created_at       VARCHAR(50)  NOT NULL
);

-- ---------------------------------------------------------------------------
-- ENTITY: production_order_items  (ProductionItem inside ProductionOrder)
-- Items to be produced within a production order.
-- Relationship: N:1 with production_orders
--               N:1 with products
--               1:N with production_item_material_requirements
-- ---------------------------------------------------------------------------
CREATE TABLE production_order_items (
    id                  UUID   NOT NULL DEFAULT gen_random_uuid() PRIMARY KEY,
    production_order_id UUID   NOT NULL REFERENCES production_orders(id) ON DELETE CASCADE,
    product_id          UUID   NOT NULL REFERENCES products(id)          ON DELETE NO ACTION,
    quantity            BIGINT NOT NULL CHECK (quantity > 0)
);

-- ---------------------------------------------------------------------------
-- VALUE OBJECT: production_item_material_requirements  (MaterialRequirement inside ProductionItem)
-- Dry supply amounts needed for a specific production item.
-- Relationship: N:1 with production_order_items
--               N:1 with dry_supplies
-- ---------------------------------------------------------------------------
CREATE TABLE production_item_material_requirements (
    id                       UUID   NOT NULL DEFAULT gen_random_uuid() PRIMARY KEY,
    production_order_item_id UUID   NOT NULL REFERENCES production_order_items(id) ON DELETE CASCADE,
    dry_supply_id            UUID   NOT NULL REFERENCES dry_supplies(id)            ON DELETE NO ACTION,
    quantity                 BIGINT NOT NULL CHECK (quantity > 0)
);

-- ---------------------------------------------------------------------------
-- AGGREGATE: dry_supply_inventories
-- Tracks stock for a single dry supply (three-layer model):
--   Physical  = IN - CONSUMED +/- ADJUSTED
--   Committed = COMMITTED - RELEASED - CONSUMED
--   Available = Physical - Committed
-- Relationship: 1:1 with dry_supplies
--               1:N with dry_supply_inventory_movements
-- ---------------------------------------------------------------------------
CREATE TABLE dry_supply_inventories (
    id            UUID NOT NULL DEFAULT gen_random_uuid() PRIMARY KEY,
    dry_supply_id UUID NOT NULL UNIQUE REFERENCES dry_supplies(id) ON DELETE NO ACTION
);

-- ---------------------------------------------------------------------------
-- VALUE OBJECT: dry_supply_inventory_movements  (DrySupplyMovement inside DrySupplyInventory)
-- Immutable ledger of dry supply stock changes.
-- Relationship: N:1 with dry_supply_inventories
-- ---------------------------------------------------------------------------
CREATE TABLE dry_supply_inventory_movements (
    id                      UUID         NOT NULL DEFAULT gen_random_uuid() PRIMARY KEY,
    dry_supply_inventory_id UUID         NOT NULL REFERENCES dry_supply_inventories(id) ON DELETE CASCADE,
    -- valores: DRY_SUPPLY_IN, DRY_SUPPLY_COMMITTED, DRY_SUPPLY_RELEASED, DRY_SUPPLY_CONSUMED, DRY_SUPPLY_ADJUSTED
    movement_type           VARCHAR(50)  NOT NULL,
    quantity                BIGINT       NOT NULL CHECK (quantity > 0),
    -- production order ID or purchase reference
    reference               VARCHAR(255),
    -- fecha en formato RFC3339
    created_at              VARCHAR(50)  NOT NULL
);

-- ---------------------------------------------------------------------------
-- AGGREGATE: product_inventories
-- Tracks stock per product+SKU in two stages:
--   UNDRESSED (Sin Vestir / SV): bottled but not yet labeled
--   DRESSED   (Producto Terminado / PT): fully labeled, ready to dispatch
-- Relationship: N:1 with products (one record per product+SKU combination)
--               1:N with product_inventory_movements
-- ---------------------------------------------------------------------------
CREATE TABLE product_inventories (
    id         UUID         NOT NULL DEFAULT gen_random_uuid() PRIMARY KEY,
    product_id UUID         NOT NULL REFERENCES products(id) ON DELETE NO ACTION,
    sku        VARCHAR(100) NOT NULL,
    UNIQUE (product_id, sku)
);

-- ---------------------------------------------------------------------------
-- VALUE OBJECT: product_inventory_movements  (ProductMovement inside ProductInventory)
-- Immutable ledger of product stock changes across both stages.
-- Relationship: N:1 with product_inventories
-- ---------------------------------------------------------------------------
CREATE TABLE product_inventory_movements (
    id                   UUID         NOT NULL DEFAULT gen_random_uuid() PRIMARY KEY,
    product_inventory_id UUID         NOT NULL REFERENCES product_inventories(id) ON DELETE CASCADE,
    -- valores: PRODUCT_PRODUCED, PRODUCT_STAGE_IN, PRODUCT_STAGE_OUT,
    --          PRODUCT_RESERVED_FOR_SALE, PRODUCT_RESERVATION_RELEASED,
    --          PRODUCT_DISPATCHED, PRODUCT_STOCK_ADJUSTED
    movement_type        VARCHAR(50)  NOT NULL,
    quantity             BIGINT       NOT NULL CHECK (quantity > 0),
    -- valores: DRESSED, UNDRESSED
    stage                VARCHAR(20)  NOT NULL,
    -- saleOrderID, productionOrderID or dispatchID
    reference            VARCHAR(255) NOT NULL,
    -- ej: L-081124-38-11, asignado al convertir SV a PT
    lot_number           VARCHAR(100),
    -- fecha en formato RFC3339
    created_at           VARCHAR(50)  NOT NULL
);

-- ---------------------------------------------------------------------------
-- ADMINISTRATIVE PLAN V1A: customer invoicing, remittances and receipts
-- Additive/idempotent schema shared with runtime migrations.
-- ---------------------------------------------------------------------------
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

-- ---------------------------------------------------------------------------
-- ADMINISTRATIVE PLAN V1B: suppliers, purchasing, invoices and payments
-- ---------------------------------------------------------------------------
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

-- ---------------------------------------------------------------------------
-- INDEXES
-- ---------------------------------------------------------------------------
CREATE INDEX idx_sale_orders_customer_id         ON sale_orders(customer_id);
CREATE INDEX idx_sale_orders_status              ON sale_orders(status);
CREATE INDEX idx_sale_order_items_sale_order_id  ON sale_order_items(sale_order_id);
CREATE INDEX idx_sale_order_items_product_id     ON sale_order_items(product_id);
CREATE INDEX idx_production_orders_sale_order_id ON production_orders(sale_order_id);
CREATE INDEX idx_prod_order_items_prod_order_id  ON production_order_items(production_order_id);
CREATE INDEX idx_prod_item_reqs_item_id          ON production_item_material_requirements(production_order_item_id);
CREATE INDEX idx_ds_inv_movements_inventory_id   ON dry_supply_inventory_movements(dry_supply_inventory_id);
CREATE INDEX idx_prod_inv_movements_inventory_id ON product_inventory_movements(product_inventory_id);
CREATE INDEX idx_product_bods_product_id         ON product_bill_of_dry_supply(product_id);

-- Administrative plan V1A/V1B indexes. Unique constraints above cover
-- idempotency keys, document numbering and item line numbering.
CREATE INDEX IF NOT EXISTS idx_sales_invoices_customer_id              ON sales_invoices(customer_id);
CREATE INDEX IF NOT EXISTS idx_sales_invoices_sale_order_id            ON sales_invoices(sale_order_id);
CREATE INDEX IF NOT EXISTS idx_sales_invoices_issue_date               ON sales_invoices(issue_date);
CREATE INDEX IF NOT EXISTS idx_sales_invoices_status                   ON sales_invoices(status);
CREATE INDEX IF NOT EXISTS idx_sales_invoice_items_invoice_id          ON sales_invoice_items(sales_invoice_id);
CREATE INDEX IF NOT EXISTS idx_sales_invoice_items_product_id          ON sales_invoice_items(product_id);
CREATE INDEX IF NOT EXISTS idx_remittances_customer_id                 ON remittances(customer_id);
CREATE INDEX IF NOT EXISTS idx_remittances_sale_order_id               ON remittances(sale_order_id);
CREATE INDEX IF NOT EXISTS idx_remittances_sales_invoice_id            ON remittances(sales_invoice_id);
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

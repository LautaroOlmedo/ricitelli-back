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

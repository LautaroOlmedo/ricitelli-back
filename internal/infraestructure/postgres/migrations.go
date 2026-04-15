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
	}

	for _, m := range migrations {
		if _, err := pool.Exec(context.Background(), m.sql); err != nil {
			return err
		}
		log.Printf("[migration] applied: %s\n", m.name)
	}
	return nil
}

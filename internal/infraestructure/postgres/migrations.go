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
	}

	for _, m := range migrations {
		if _, err := pool.Exec(context.Background(), m.sql); err != nil {
			return err
		}
		log.Printf("[migration] applied: %s\n", m.name)
	}
	return nil
}

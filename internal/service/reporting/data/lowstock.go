package data

import (
	"context"
	"sort"
	"time"

	inventory_svc "ricitelli-back/internal/service/inventory"
)

type LowStockSource interface {
	GetLowStockAlerts(ctx context.Context) ([]inventory_svc.DrySupplyAlert, error)
	GetMovements(ctx context.Context, filter inventory_svc.MovementFilter) (*inventory_svc.MovementsResult, error)
}

// LowStockItem enriches an alert with a recommended reorder quantity.
type LowStockItem struct {
	SupplyID           string
	Code               string
	Name               string
	Physical           int64
	Committed          int64
	Available          int64
	ReorderPoint       int64  // not directly exposed — inferred as threshold that made it low
	AvgDailyConsumption float64
	RecommendedReorder int64 // avgDaily * 30 * safety
	Criticality        float64 // 1 - available/reorderPoint clamped 0..1
}

// AggregateLowStock returns enriched low-stock items.
func AggregateLowStock(ctx context.Context, src LowStockSource) ([]LowStockItem, error) {
	alerts, err := src.GetLowStockAlerts(ctx)
	if err != nil {
		return nil, err
	}
	if len(alerts) == 0 {
		return nil, nil
	}

	// Pull last 30 days of consumption to compute averages.
	from := time.Now().Add(-30 * 24 * time.Hour).UTC().Format(time.RFC3339)
	to := time.Now().UTC().Format(time.RFC3339)
	mov, err := src.GetMovements(ctx, inventory_svc.MovementFilter{
		FromDate:     from,
		ToDate:       to,
		Category:     "DRY_SUPPLY",
		MovementType: "DRY_SUPPLY_CONSUMED",
		PageSize:     100000,
	})
	consumedBySupply := map[string]uint64{}
	if err == nil && mov != nil {
		for _, m := range mov.Movements {
			consumedBySupply[m.ItemName] += m.Quantity
		}
	}

	items := make([]LowStockItem, 0, len(alerts))
	for _, a := range alerts {
		if !a.IsLow {
			continue
		}
		avgDaily := float64(consumedBySupply[a.Name]) / 30.0
		recommended := int64(avgDaily * 30 * 2) // 2x safety factor -> 60 days of stock
		// Fall back to reaching 3x current committed if no history
		if recommended == 0 {
			recommended = a.Committed*3 + 500
		}
		// Derive reorderPoint as max(available+1, 500) — alerts were filtered by <500 in service
		// Best effort: use 500 as a safe default if not exposed.
		rp := int64(500)
		var crit float64
		if rp > 0 {
			crit = 1.0 - float64(a.Available)/float64(rp)
			if crit < 0 {
				crit = 0
			}
			if crit > 1 {
				crit = 1
			}
		}
		items = append(items, LowStockItem{
			SupplyID:            a.DrySupplyID,
			Code:                a.Code,
			Name:                a.Name,
			Physical:            a.Physical,
			Committed:           a.Committed,
			Available:           a.Available,
			ReorderPoint:        rp,
			AvgDailyConsumption: avgDaily,
			RecommendedReorder:  recommended,
			Criticality:         crit,
		})
	}

	sort.Slice(items, func(i, j int) bool {
		return items[i].Criticality > items[j].Criticality
	})
	return items, nil
}

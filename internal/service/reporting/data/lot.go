package data

import (
	"context"

	inventory_svc "ricitelli-back/internal/service/inventory"
)

// LotTrace reconstructs the movement chain for a lot_number.
type LotTrace struct {
	LotNumber  string
	Movements  []inventory_svc.MovementEntry
	ProductName string
	TotalUnits uint64
}

// AggregateLotTrace returns all movements tied to the given lot_number.
// We scan the full movement log (page_size large) and filter client-side since
// current repo doesn't expose a lot_number filter yet.
func AggregateLotTrace(ctx context.Context, inv InventoryMovementsSource, lotNumber string) (*LotTrace, error) {
	res, err := inv.GetMovements(ctx, inventory_svc.MovementFilter{
		Category: "PRODUCT",
		PageSize: 100000,
	})
	if err != nil {
		return nil, err
	}
	trace := &LotTrace{LotNumber: lotNumber}
	for _, m := range res.Movements {
		if m.LotNumber == lotNumber {
			trace.Movements = append(trace.Movements, m)
			if trace.ProductName == "" {
				trace.ProductName = m.ItemName
			}
			if m.MovementType == "PRODUCT_STAGE_IN" && m.Stage == "DRESSED" {
				trace.TotalUnits += m.Quantity
			}
		}
	}
	return trace, nil
}

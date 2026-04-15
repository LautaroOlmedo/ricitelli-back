package data

import (
	"context"
	"sort"
	"time"

	production_order_domain "ricitelli-back/internal/domain/production-order"
	inventory_svc "ricitelli-back/internal/service/inventory"
)

type ProductionOrderSource interface {
	GetProductionOrders(ctx context.Context) ([]production_order_domain.ProductionOrder, error)
}

type InventoryMovementsSource interface {
	GetMovements(ctx context.Context, filter inventory_svc.MovementFilter) (*inventory_svc.MovementsResult, error)
}

type DrySupplyLookup interface {
	GetDrySupplyCategory(ctx context.Context, id string) string
	GetDrySupplyName(ctx context.Context, id string) string
}

// ProductionAggregates holds the metrics for the production report.
type ProductionAggregates struct {
	SVProduced         uint64
	SVtoPTConverted    uint64
	Dispatched         uint64
	SuppliesIn         uint64
	SuppliesConsumed   uint64
	POCompleted        int
	POInProgress       int
	POCancelled        int

	SVvsPTByProduct       []StackedProduct
	ConversionsByDay      []DayValue
	SuppliesInVsConsByDay []PairedDayValue
	TopConsumedSupplies   []NameValue
	ConsumptionByCategory map[string]float64
	LotsByDay             []DayValue
	POStatus              map[string]int

	RecentMovements []inventory_svc.MovementEntry
}

// StackedProduct carries SV and PT values per product name.
type StackedProduct struct {
	ProductName string
	SV          float64
	PT          float64
}

// PairedDayValue holds two series for the same day.
type PairedDayValue struct {
	Day  time.Time
	A, B float64
}

// AggregateProduction builds metrics from movements and production orders.
func AggregateProduction(
	ctx context.Context,
	inv InventoryMovementsSource,
	po ProductionOrderSource,
	from, to string,
) (*ProductionAggregates, error) {
	// Pull all movements in range (big page size).
	res, err := inv.GetMovements(ctx, inventory_svc.MovementFilter{
		FromDate: from,
		ToDate:   to,
		Page:     1,
		PageSize: 100000,
	})
	if err != nil {
		return nil, err
	}

	agg := &ProductionAggregates{
		ConsumptionByCategory: map[string]float64{},
		POStatus:              map[string]int{},
	}

	svPerProduct := map[string]float64{}
	ptPerProduct := map[string]float64{}
	convByDay := map[string]float64{}
	suppliesInByDay := map[string]float64{}
	suppliesConsByDay := map[string]float64{}
	consumedBySupply := map[string]float64{}
	lotsByDay := map[string]map[string]struct{}{}

	for _, m := range res.Movements {
		day := ""
		if t, err := time.Parse(time.RFC3339, m.CreatedAt); err == nil {
			day = t.Format("2006-01-02")
		}
		switch m.MovementType {
		case "PRODUCT_PRODUCED":
			agg.SVProduced += m.Quantity
			svPerProduct[m.ItemName] += float64(m.Quantity)
		case "PRODUCT_STAGE_IN":
			if m.Stage == "DRESSED" {
				agg.SVtoPTConverted += m.Quantity
				ptPerProduct[m.ItemName] += float64(m.Quantity)
				if day != "" {
					convByDay[day] += float64(m.Quantity)
				}
				if m.LotNumber != "" && day != "" {
					if _, ok := lotsByDay[day]; !ok {
						lotsByDay[day] = map[string]struct{}{}
					}
					lotsByDay[day][m.LotNumber] = struct{}{}
				}
			} else if m.Stage == "UNDRESSED" {
				// sometimes used for reconciling SV
				svPerProduct[m.ItemName] += float64(m.Quantity)
			}
		case "PRODUCT_DISPATCHED":
			agg.Dispatched += m.Quantity
		case "DRY_SUPPLY_IN":
			agg.SuppliesIn += m.Quantity
			if day != "" {
				suppliesInByDay[day] += float64(m.Quantity)
			}
		case "DRY_SUPPLY_CONSUMED":
			agg.SuppliesConsumed += m.Quantity
			consumedBySupply[m.ItemName] += float64(m.Quantity)
			if day != "" {
				suppliesConsByDay[day] += float64(m.Quantity)
			}
		}

		// Keep recent 30 for the summary table
		agg.RecentMovements = append(agg.RecentMovements, m)
	}
	// Recent = first 30 of result (already DESC by created_at from repo)
	if len(agg.RecentMovements) > 30 {
		agg.RecentMovements = agg.RecentMovements[:30]
	}

	// Production orders by status + filter by date in range
	orders, err := po.GetProductionOrders(ctx)
	if err == nil {
		fromT, _ := time.Parse(time.RFC3339, from)
		toT, _ := time.Parse(time.RFC3339, to)
		for _, o := range orders {
			created, err := time.Parse(time.RFC3339, o.GetCreatedAt())
			if err != nil {
				continue
			}
			if !fromT.IsZero() && created.Before(fromT) {
				continue
			}
			if !toT.IsZero() && !created.Before(toT) {
				continue
			}
			status := string(o.GetStatus())
			agg.POStatus[status]++
			switch o.GetStatus() {
			case production_order_domain.StatusCompleted:
				agg.POCompleted++
			case production_order_domain.StatusInProgress:
				agg.POInProgress++
			case production_order_domain.StatusCancelled:
				agg.POCancelled++
			}
		}
	}

	// Compose SVvsPTByProduct
	products := map[string]struct{}{}
	for k := range svPerProduct {
		products[k] = struct{}{}
	}
	for k := range ptPerProduct {
		products[k] = struct{}{}
	}
	for name := range products {
		agg.SVvsPTByProduct = append(agg.SVvsPTByProduct, StackedProduct{
			ProductName: name,
			SV:          svPerProduct[name],
			PT:          ptPerProduct[name],
		})
	}
	sort.Slice(agg.SVvsPTByProduct, func(i, j int) bool {
		return agg.SVvsPTByProduct[i].SV+agg.SVvsPTByProduct[i].PT >
			agg.SVvsPTByProduct[j].SV+agg.SVvsPTByProduct[j].PT
	})

	// Days series
	agg.ConversionsByDay = orderedDays(convByDay)
	agg.SuppliesInVsConsByDay = orderedPairedDays(suppliesInByDay, suppliesConsByDay)
	agg.TopConsumedSupplies = topN(consumedBySupply, 10)

	// lots per day
	lotByDayFloat := map[string]float64{}
	for day, set := range lotsByDay {
		lotByDayFloat[day] = float64(len(set))
	}
	agg.LotsByDay = orderedDays(lotByDayFloat)

	return agg, nil
}

func orderedDays(m map[string]float64) []DayValue {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	out := make([]DayValue, 0, len(keys))
	for _, k := range keys {
		t, _ := time.Parse("2006-01-02", k)
		out = append(out, DayValue{Day: t, Value: m[k]})
	}
	return out
}

func orderedPairedDays(a, b map[string]float64) []PairedDayValue {
	keys := map[string]struct{}{}
	for k := range a {
		keys[k] = struct{}{}
	}
	for k := range b {
		keys[k] = struct{}{}
	}
	ks := make([]string, 0, len(keys))
	for k := range keys {
		ks = append(ks, k)
	}
	sort.Strings(ks)
	out := make([]PairedDayValue, 0, len(ks))
	for _, k := range ks {
		t, _ := time.Parse("2006-01-02", k)
		out = append(out, PairedDayValue{Day: t, A: a[k], B: b[k]})
	}
	return out
}

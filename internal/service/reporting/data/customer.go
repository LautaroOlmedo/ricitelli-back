package data

import (
	"context"
	"sort"
	"time"

	customer_domain "ricitelli-back/internal/domain/customer"
	sale_order_domain "ricitelli-back/internal/domain/sale-order"
)

// CustomerAggregates holds per-customer summary data.
type CustomerAggregates struct {
	Customer          *customer_domain.Customer
	Orders            []sale_order_domain.SaleOrder
	TotalRevenueARS   float64
	OrderCount        int
	AverageTicket     float64
	FirstPurchase     string
	LastPurchase      string
	RevenueByMonth    []DayValue
	TopProducts       []NameValue
	RevenueByCurrency map[string]float64
}

// AggregateCustomer computes all metrics for a single customer in a date range.
func AggregateCustomer(
	ctx context.Context,
	orders SaleOrderSource,
	customers CustomerSource,
	products ProductSource,
	customerID, from, to string,
) (*CustomerAggregates, error) {
	// Normalize RFC3339 -> YYYY-MM-DD for storage; keep originals for fine filter.
	fromDay, toDay := from, to
	fromTS, _ := time.Parse(time.RFC3339, from)
	toTS, _ := time.Parse(time.RFC3339, to)
	if !fromTS.IsZero() {
		fromDay = fromTS.UTC().Format("2006-01-02")
	}
	if !toTS.IsZero() {
		toDay = toTS.UTC().Format("2006-01-02")
	}

	var all []sale_order_domain.SaleOrder
	if from != "" && to != "" {
		o, err := orders.GetSaleOrdersByDateRange(ctx, fromDay, toDay)
		if err != nil {
			return nil, err
		}
		all = o
	}

	c, _ := customers.GetCustomerByID(ctx, customerID)
	agg := &CustomerAggregates{
		Customer:          c,
		RevenueByCurrency: map[string]float64{},
	}

	revByMonth := map[string]float64{}
	revByProduct := map[string]float64{}
	for _, o := range all {
		if o.GetCustomerID() != customerID {
			continue
		}
		// Hour-precision filter using original RFC3339 bounds.
		if ct, err := time.Parse(time.RFC3339, o.GetCreatedAt()); err == nil {
			if !fromTS.IsZero() && ct.Before(fromTS) {
				continue
			}
			if !toTS.IsZero() && ct.After(toTS) {
				continue
			}
		}
		agg.Orders = append(agg.Orders, o)
		var amount float64
		for _, it := range o.GetItems() {
			line := float64(it.Quantity) * float64(it.UnitPrice)
			amount += line
			name := it.ProductID
			if p, err := products.GetProductByID(ctx, it.ProductID); err == nil && p != nil {
				name = p.GetName()
			}
			revByProduct[name] += line
		}
		currency := string(o.GetCurrency())
		agg.RevenueByCurrency[currency] += amount
		if currency == "ARS" {
			agg.TotalRevenueARS += amount
		}
		if t, err := time.Parse(time.RFC3339, o.GetCreatedAt()); err == nil {
			m := t.Format("2006-01")
			if currency == "ARS" {
				revByMonth[m] += amount
			}
			if agg.FirstPurchase == "" || o.GetCreatedAt() < agg.FirstPurchase {
				agg.FirstPurchase = o.GetCreatedAt()
			}
			if o.GetCreatedAt() > agg.LastPurchase {
				agg.LastPurchase = o.GetCreatedAt()
			}
		}
	}

	agg.OrderCount = len(agg.Orders)
	if agg.OrderCount > 0 {
		agg.AverageTicket = agg.TotalRevenueARS / float64(agg.OrderCount)
	}

	keys := make([]string, 0, len(revByMonth))
	for k := range revByMonth {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		t, _ := time.Parse("2006-01", k)
		agg.RevenueByMonth = append(agg.RevenueByMonth, DayValue{Day: t, Value: revByMonth[k]})
	}
	agg.TopProducts = topN(revByProduct, 10)
	return agg, nil
}

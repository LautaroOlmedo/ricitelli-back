package data

import (
	"context"
	"sort"
	"time"

	customer_domain "ricitelli-back/internal/domain/customer"
	product_domain "ricitelli-back/internal/domain/product"
	sale_order_domain "ricitelli-back/internal/domain/sale-order"
)

// SaleOrderSource is the abstraction used by the sales aggregator.
type SaleOrderSource interface {
	GetSaleOrdersByDateRange(ctx context.Context, from, to string) ([]sale_order_domain.SaleOrder, error)
}

type CustomerSource interface {
	GetCustomerByID(ctx context.Context, id string) (*customer_domain.Customer, error)
}

type ProductSource interface {
	GetProductByID(ctx context.Context, id string) (*product_domain.Product, error)
}

// SalesAggregates is the full data set used to render the sales report.
type SalesAggregates struct {
	Orders               []sale_order_domain.SaleOrder
	Customers            map[string]*customer_domain.Customer
	Products             map[string]*product_domain.Product
	TotalRevenueARS      float64
	OrderCount           int
	UnitsDispatched      uint64
	AverageTicket        float64
	FulfillmentRate      float64 // dispatched / total
	RevenueByDay         []DayValue
	OrdersByMarket       map[string]int
	RevenueByCurrency    map[string]float64
	TopProductsByRevenue []NameValue
	SalesByCustomerGroup map[string]int
	OrdersByCountry      map[string]int
	StatusDistribution   map[string]int
	TopOrders            []TopOrder
}

type DayValue struct {
	Day   time.Time
	Value float64
}

type NameValue struct {
	Name  string
	Value float64
}

type TopOrder struct {
	CreatedAt  string
	CustomerID string
	CustomerName string
	ProductName  string
	Amount     float64
	Currency   string
	Status     string
}

// SalesFilter narrows the order set beyond date range.
type SalesFilter struct {
	Market     string
	Currency   string
	CustomerID string
	ProductID  string
}

// AggregateSales builds all the metrics needed by the sales report.
func AggregateSales(
	ctx context.Context,
	orders SaleOrderSource,
	customers CustomerSource,
	products ProductSource,
	from, to string,
	filter SalesFilter,
) (*SalesAggregates, error) {
	// Storage layer expects YYYY-MM-DD; convert from RFC3339 and keep originals
	// for fine-grained hour filtering below.
	fromDay, toDay := from, to
	fromTS, _ := time.Parse(time.RFC3339, from)
	toTS, _ := time.Parse(time.RFC3339, to)
	if !fromTS.IsZero() {
		fromDay = fromTS.UTC().Format("2006-01-02")
	}
	if !toTS.IsZero() {
		toDay = toTS.UTC().Format("2006-01-02")
	}

	all, err := orders.GetSaleOrdersByDateRange(ctx, fromDay, toDay)
	if err != nil {
		return nil, err
	}

	agg := &SalesAggregates{
		Customers:            map[string]*customer_domain.Customer{},
		Products:             map[string]*product_domain.Product{},
		OrdersByMarket:       map[string]int{},
		RevenueByCurrency:    map[string]float64{},
		SalesByCustomerGroup: map[string]int{},
		OrdersByCountry:      map[string]int{},
		StatusDistribution:   map[string]int{},
	}

	revenueByDay := map[string]float64{}
	revenueByProduct := map[string]float64{}

	for _, o := range all {
		// Hour-precision filter using original RFC3339 bounds.
		if ct, err := time.Parse(time.RFC3339, o.GetCreatedAt()); err == nil {
			if !fromTS.IsZero() && ct.Before(fromTS) {
				continue
			}
			if !toTS.IsZero() && ct.After(toTS) {
				continue
			}
		}
		if filter.Market != "" && string(o.GetMarket()) != filter.Market {
			continue
		}
		if filter.Currency != "" && string(o.GetCurrency()) != filter.Currency {
			continue
		}
		if filter.CustomerID != "" && o.GetCustomerID() != filter.CustomerID {
			continue
		}
		if filter.ProductID != "" {
			found := false
			for _, it := range o.GetItems() {
				if it.ProductID == filter.ProductID {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}

		agg.Orders = append(agg.Orders, o)

		// Customer
		custName := o.GetCustomerID()
		custGroup := "UNKNOWN"
		if _, ok := agg.Customers[o.GetCustomerID()]; !ok {
			if c, err := customers.GetCustomerByID(ctx, o.GetCustomerID()); err == nil && c != nil {
				agg.Customers[o.GetCustomerID()] = c
				custName = c.GetSocialReason()
				custGroup = string(c.GetGroup())
			}
		} else {
			c := agg.Customers[o.GetCustomerID()]
			custName = c.GetSocialReason()
			custGroup = string(c.GetGroup())
		}

		// Order total
		var orderAmount float64
		var firstProductName string
		for _, it := range o.GetItems() {
			lineAmount := float64(it.Quantity) * float64(it.UnitPrice)
			orderAmount += lineAmount

			// Product
			pName := it.ProductID
			if _, ok := agg.Products[it.ProductID]; !ok {
				if p, err := products.GetProductByID(ctx, it.ProductID); err == nil && p != nil {
					agg.Products[it.ProductID] = p
					pName = p.GetName()
				}
			} else {
				pName = agg.Products[it.ProductID].GetName()
			}
			if firstProductName == "" {
				firstProductName = pName
			}
			revenueByProduct[pName] += lineAmount
		}

		currency := string(o.GetCurrency())
		agg.RevenueByCurrency[currency] += orderAmount

		// ARS equivalent — for simplicity we only sum ARS revenue. Others listed separately.
		if currency == "ARS" {
			agg.TotalRevenueARS += orderAmount
		}

		// Day bucket
		if t, err := time.Parse(time.RFC3339, o.GetCreatedAt()); err == nil {
			dayKey := t.Format("2006-01-02")
			if currency == "ARS" {
				revenueByDay[dayKey] += orderAmount
			}
		}

		// Market / country / group / status
		agg.OrdersByMarket[string(o.GetMarket())]++
		if o.GetDestinationCountry() != "" {
			agg.OrdersByCountry[o.GetDestinationCountry()]++
		}
		agg.SalesByCustomerGroup[custGroup]++
		agg.StatusDistribution[string(o.GetStatus())]++

		// Units dispatched
		if o.GetStatus() == sale_order_domain.StatusDispatched {
			for _, it := range o.GetItems() {
				agg.UnitsDispatched += it.Quantity
			}
		}

		// Top orders — collect up to 20 later
		agg.TopOrders = append(agg.TopOrders, TopOrder{
			CreatedAt:    o.GetCreatedAt(),
			CustomerID:   o.GetCustomerID(),
			CustomerName: custName,
			ProductName:  firstProductName,
			Amount:       orderAmount,
			Currency:     currency,
			Status:       string(o.GetStatus()),
		})
	}

	// Finalize aggregates
	agg.OrderCount = len(agg.Orders)
	if agg.OrderCount > 0 {
		agg.AverageTicket = agg.TotalRevenueARS / float64(agg.OrderCount)
	}
	dispatched := agg.StatusDistribution[string(sale_order_domain.StatusDispatched)]
	if agg.OrderCount > 0 {
		agg.FulfillmentRate = float64(dispatched) / float64(agg.OrderCount) * 100
	}

	// Sort days
	keys := make([]string, 0, len(revenueByDay))
	for k := range revenueByDay {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		t, _ := time.Parse("2006-01-02", k)
		agg.RevenueByDay = append(agg.RevenueByDay, DayValue{Day: t, Value: revenueByDay[k]})
	}

	// Top products by revenue (top 10)
	agg.TopProductsByRevenue = topN(revenueByProduct, 10)

	// Sort top orders by amount desc, keep 20
	sort.Slice(agg.TopOrders, func(i, j int) bool {
		return agg.TopOrders[i].Amount > agg.TopOrders[j].Amount
	})
	if len(agg.TopOrders) > 20 {
		agg.TopOrders = agg.TopOrders[:20]
	}

	return agg, nil
}

func topN(m map[string]float64, n int) []NameValue {
	out := make([]NameValue, 0, len(m))
	for k, v := range m {
		out = append(out, NameValue{Name: k, Value: v})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Value > out[j].Value })
	if len(out) > n {
		out = out[:n]
	}
	return out
}

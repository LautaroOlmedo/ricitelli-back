package customer

import (
	"context"
	"errors"
	"fmt"

	customer_domain "ricitelli-back/internal/domain/customer"
	sale_order "ricitelli-back/internal/domain/sale-order"
)

// PlaceOrder creates a SaleOrder on behalf of the customer.
// The Market is derived from the customer's MarketType:
//   - INTERNAL → DOMESTIC
//   - EXTERNAL → EXPORT
func (s *Service) PlaceOrder(ctx context.Context, params PlaceOrderParams) (sale_order.SaleOrder, error) {
	if params.CustomerID == "" {
		return sale_order.SaleOrder{}, errors.New("customer_id is required")
	}
	if len(params.Items) == 0 {
		return sale_order.SaleOrder{}, errors.New("order must contain at least one item")
	}

	c, err := s.Storage.GetCustomerByID(ctx, params.CustomerID)
	if err != nil {
		return sale_order.SaleOrder{}, fmt.Errorf("customer not found: %w", err)
	}
	if !c.IsActive() {
		return sale_order.SaleOrder{}, errors.New("cannot place order for an inactive customer")
	}

	market := marketFromCustomer(c)

	return s.SaleOrderStorage.CreateSaleOrder(ctx, sale_order.NewSaleOrderParams{
		CustomerID:         params.CustomerID,
		Items:              params.Items,
		Currency:           params.Currency,
		Market:             market,
		DestinationCountry: params.DestinationCountry,
		SaleType:           params.SaleType,
	})
}

// GetOrdersByCustomer returns all sale orders belonging to the given customer.
func (s *Service) GetOrdersByCustomer(ctx context.Context, customerID string) ([]sale_order.SaleOrder, error) {
	if customerID == "" {
		return nil, errors.New("customer_id is required")
	}
	if _, err := s.Storage.GetCustomerByID(ctx, customerID); err != nil {
		return nil, fmt.Errorf("customer not found: %w", err)
	}

	all, err := s.SaleOrderStorage.GetSaleOrders(ctx)
	if err != nil {
		return nil, err
	}

	var result []sale_order.SaleOrder
	for _, o := range all {
		if o.GetCustomerID() == customerID {
			result = append(result, o)
		}
	}
	return result, nil
}

func marketFromCustomer(c *customer_domain.Customer) sale_order.Market {
	if c.GetMarketType() == customer_domain.MarketTypeExternal {
		return sale_order.MarketExport
	}
	return sale_order.MarketDomestic
}

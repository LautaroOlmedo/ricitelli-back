package sale_order

import (
	"errors"
	"time"

	valueObject "ricitelli-back/internal/value-object"

	"github.com/google/uuid"
)

// Status pipeline: NEW → CONFIRMED → INVOICED → DISPATCHED (or CANCELLED)
type Status string

const (
	StatusNew             Status = "NEW"
	StatusReadyToDispatch Status = "READY_TO_DISPATCH"
	StatusConfirmed       Status = "CONFIRMED"
	StatusInvoiced        Status = "INVOICED"
	StatusDispatched      Status = "DISPATCHED"
	StatusCancelled       Status = "CANCELLED"
)

// Currency supported currencies
type Currency string

const (
	CurrencyARS Currency = "ARS" // Argentine Peso
	CurrencyUSD Currency = "USD" // US Dollar
	CurrencyCAD Currency = "CAD" // Canadian Dollar
	CurrencyEUR Currency = "EUR" // Euro (UK, EU exports)
)

// Market differentiates domestic vs export orders
type Market string

const (
	MarketDomestic Market = "DOMESTIC"
	MarketExport   Market = "EXPORT"
)

// SaleType classifies the exit reason (for non-distorting commercial revenue)
type SaleType string

const (
	SaleTypeRegular          SaleType = "SALE"              // regular commercial sale
	SaleTypeSampleCustoms    SaleType = "SAMPLE_CUSTOMS"    // muestra de aduana
	SaleTypeGift             SaleType = "GIFT"              // obsequio bodega
	SaleTypeInternal         SaleType = "INTERNAL"          // consumidor final empleados
	SaleTypeCommercialSample SaleType = "COMMERCIAL_SAMPLE" // muestras comerciales (prensa/degustación)
)

// SaleOrder represents the aggregate root for a customer sales order.
type SaleOrder struct {
	id                 string
	customerID         string
	status             Status
	items              []valueObject.SaleOrderItem
	currency           Currency
	market             Market
	destinationCountry string // ISO-3166 alpha-2, e.g., "AR", "GB", "BR", "JP"
	saleType           SaleType
	createdAt          string
	active             bool
}

type NewSaleOrderParams struct {
	CustomerID         string
	Items              []valueObject.SaleOrderItem
	Currency           Currency
	Market             Market
	DestinationCountry string
	SaleType           SaleType
}

func NewSaleOrder(params NewSaleOrderParams) (SaleOrder, error) {
	if params.CustomerID == "" {
		return SaleOrder{}, errors.New("customerID cannot be empty")
	}
	if len(params.Items) == 0 {
		return SaleOrder{}, errors.New("sale order must have at least one item")
	}
	currency := params.Currency
	if currency == "" {
		currency = CurrencyARS
	}
	market := params.Market
	if market == "" {
		market = MarketDomestic
	}
	saleType := params.SaleType
	if saleType == "" {
		saleType = SaleTypeRegular
	}
	return SaleOrder{
		id:                 uuid.New().String(),
		customerID:         params.CustomerID,
		status:             StatusNew,
		items:              params.Items,
		currency:           currency,
		market:             market,
		destinationCountry: params.DestinationCountry,
		saleType:           saleType,
		createdAt:          time.Now().UTC().Format(time.RFC3339),
		active:             true,
	}, nil
}

// validTransitions maps allowed status progressions
var validTransitions = map[Status][]Status{
	StatusNew:             {StatusReadyToDispatch, StatusCancelled},
	StatusReadyToDispatch: {StatusInvoiced, StatusCancelled},
	StatusConfirmed:       {StatusInvoiced, StatusCancelled},
	StatusInvoiced:        {StatusDispatched, StatusCancelled},
}

// UpdateStatus advances the order through its lifecycle pipeline.
func (s *SaleOrder) UpdateStatus(newStatus Status) error {
	allowed, ok := validTransitions[s.status]
	if !ok {
		return errors.New("order is in a terminal state")
	}
	for _, a := range allowed {
		if a == newStatus {
			s.status = newStatus
			return nil
		}
	}
	return errors.New("invalid status transition from " + string(s.status) + " to " + string(newStatus))
}

func (s *SaleOrder) GetID() string                 { return s.id }
func (s *SaleOrder) GetCustomerID() string         { return s.customerID }
func (s *SaleOrder) GetStatus() Status             { return s.status }
func (s *SaleOrder) GetCurrency() Currency         { return s.currency }
func (s *SaleOrder) GetMarket() Market             { return s.market }
func (s *SaleOrder) GetDestinationCountry() string { return s.destinationCountry }
func (s *SaleOrder) GetSaleType() SaleType         { return s.saleType }
func (s *SaleOrder) GetCreatedAt() string          { return s.createdAt }
func (s *SaleOrder) IsActive() bool                { return s.active }

// GetItems returns a defensive copy
func (s *SaleOrder) GetItems() []valueObject.SaleOrderItem {
	itemsCopy := make([]valueObject.SaleOrderItem, len(s.items))
	copy(itemsCopy, s.items)
	return itemsCopy
}

// ReconstitueSaleOrder reconstitutes a SaleOrder from stored data (bypasses validation).
func ReconstitueSaleOrder(id, customerID string, status Status, items []valueObject.SaleOrderItem,
	currency Currency, market Market, destinationCountry string, saleType SaleType, createdAt string, active bool) SaleOrder {
	return SaleOrder{
		id:                 id,
		customerID:         customerID,
		status:             status,
		items:              items,
		currency:           currency,
		market:             market,
		destinationCountry: destinationCountry,
		saleType:           saleType,
		createdAt:          createdAt,
		active:             active,
	}
}

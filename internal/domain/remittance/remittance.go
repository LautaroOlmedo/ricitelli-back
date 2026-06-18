package remittance

import (
	"errors"
	"time"

	valueObject "ricitelli-back/internal/value-object"

	"github.com/google/uuid"
)

var (
	ErrInvalidRemittance   = errors.New("invalid remittance")
	ErrRemittanceImmutable = errors.New("confirmed remittance is immutable")
)

type Status string

const (
	StatusDraft     Status = "DRAFT"
	StatusConfirmed Status = "CONFIRMED"
)

type Item struct {
	LineNumber  int32
	ProductID   string
	Description string
	Quantity    string
	LotNumber   string
}

type NewRemittanceParams struct {
	CustomerID     string
	SaleOrderID    string
	SalesInvoiceID string
	PointOfSale    int32
	DocumentNumber int64
	IssueDate      string
	DeliveryDate   string
	IdempotencyKey string
	Items          []Item
}

type Remittance struct {
	id             string
	customerID     string
	saleOrderID    string
	salesInvoiceID string
	pointOfSale    int32
	documentNumber int64
	issueDate      string
	deliveryDate   string
	status         Status
	idempotencyKey string
	items          []Item
	createdAt      string
	updatedAt      string
}

func NewRemittance(params NewRemittanceParams) (Remittance, error) {
	if err := validateParams(params); err != nil {
		return Remittance{}, err
	}
	now := time.Now().UTC().Format(time.RFC3339Nano)
	return Remittance{
		id: uuid.NewString(), customerID: params.CustomerID, saleOrderID: params.SaleOrderID,
		salesInvoiceID: params.SalesInvoiceID, pointOfSale: params.PointOfSale, documentNumber: params.DocumentNumber,
		issueDate: params.IssueDate, deliveryDate: params.DeliveryDate, status: StatusDraft,
		idempotencyKey: params.IdempotencyKey, items: copyItems(params.Items), createdAt: now, updatedAt: now,
	}, nil
}

func validateParams(params NewRemittanceParams) error {
	if params.CustomerID == "" || params.SaleOrderID == "" || params.PointOfSale <= 0 ||
		params.DocumentNumber <= 0 || params.IdempotencyKey == "" || len(params.Items) == 0 {
		return ErrInvalidRemittance
	}
	issue, err := time.Parse("2006-01-02", params.IssueDate)
	if err != nil {
		return ErrInvalidRemittance
	}
	if params.DeliveryDate != "" {
		delivery, err := time.Parse("2006-01-02", params.DeliveryDate)
		if err != nil || delivery.Before(issue) {
			return ErrInvalidRemittance
		}
	}
	lineNumbers := map[int32]struct{}{}
	productIDs := map[string]struct{}{}
	for _, item := range params.Items {
		if item.LineNumber <= 0 || item.ProductID == "" || item.Description == "" ||
			valueObject.ValidateDecimal(item.Quantity, 4, true) != nil {
			return ErrInvalidRemittance
		}
		if _, exists := lineNumbers[item.LineNumber]; exists {
			return ErrInvalidRemittance
		}
		lineNumbers[item.LineNumber] = struct{}{}
		if _, exists := productIDs[item.ProductID]; exists {
			return ErrInvalidRemittance
		}
		productIDs[item.ProductID] = struct{}{}
		// Current product inventory is unit-based. Fractional dispatches cannot
		// be represented and are rejected before confirmation.
		if _, err := valueObject.DecimalToUint64(item.Quantity); err != nil {
			return ErrInvalidRemittance
		}
	}
	return nil
}

// Confirm is idempotent. The returned boolean is true only for the transition
// that must cause the physical-stock dispatch.
func (r *Remittance) Confirm() (bool, error) {
	if r.status == StatusConfirmed {
		return false, nil
	}
	if r.status != StatusDraft {
		return false, ErrRemittanceImmutable
	}
	r.status = StatusConfirmed
	r.updatedAt = time.Now().UTC().Format(time.RFC3339Nano)
	return true, nil
}

func (r Remittance) GetID() string             { return r.id }
func (r Remittance) GetCustomerID() string     { return r.customerID }
func (r Remittance) GetSaleOrderID() string    { return r.saleOrderID }
func (r Remittance) GetSalesInvoiceID() string { return r.salesInvoiceID }
func (r Remittance) GetPointOfSale() int32     { return r.pointOfSale }
func (r Remittance) GetDocumentNumber() int64  { return r.documentNumber }
func (r Remittance) GetIssueDate() string      { return r.issueDate }
func (r Remittance) GetDeliveryDate() string   { return r.deliveryDate }
func (r Remittance) GetStatus() Status         { return r.status }
func (r Remittance) GetIdempotencyKey() string { return r.idempotencyKey }
func (r Remittance) GetCreatedAt() string      { return r.createdAt }
func (r Remittance) GetUpdatedAt() string      { return r.updatedAt }
func (r Remittance) IsEditable() bool          { return r.status == StatusDraft }
func (r Remittance) GetItems() []Item          { return copyItems(r.items) }

func ReconstituteRemittance(id string, params NewRemittanceParams, status Status, createdAt, updatedAt string) Remittance {
	return Remittance{
		id: id, customerID: params.CustomerID, saleOrderID: params.SaleOrderID, salesInvoiceID: params.SalesInvoiceID,
		pointOfSale: params.PointOfSale, documentNumber: params.DocumentNumber, issueDate: params.IssueDate,
		deliveryDate: params.DeliveryDate, status: status, idempotencyKey: params.IdempotencyKey,
		items: copyItems(params.Items), createdAt: createdAt, updatedAt: updatedAt,
	}
}

func copyItems(items []Item) []Item {
	result := make([]Item, len(items))
	copy(result, items)
	return result
}

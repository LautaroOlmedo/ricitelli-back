package customer_receipt

import (
	"errors"
	"time"

	valueObject "ricitelli-back/internal/value-object"

	"github.com/google/uuid"
)

var (
	ErrInvalidCustomerReceipt = errors.New("invalid customer receipt")
	ErrReceiptImmutable       = errors.New("posted customer receipt is immutable")
)

type Status string

const (
	StatusDraft  Status = "DRAFT"
	StatusPosted Status = "POSTED"
)

type Allocation struct {
	SalesInvoiceID  string
	AllocatedAmount string
}

type NewCustomerReceiptParams struct {
	CustomerID       string
	ReceiptNumber    int64
	ReceiptDate      string
	Currency         string
	ExchangeRate     string
	Amount           string
	PaymentMethod    string
	PaymentReference string
	IdempotencyKey   string
	Allocations      []Allocation
}

type CustomerReceipt struct {
	id               string
	customerID       string
	receiptNumber    int64
	receiptDate      string
	currency         string
	exchangeRate     string
	amount           string
	paymentMethod    string
	paymentReference string
	status           Status
	idempotencyKey   string
	allocations      []Allocation
	createdAt        string
	updatedAt        string
}

func NewCustomerReceipt(params NewCustomerReceiptParams) (CustomerReceipt, error) {
	if err := validateParams(params); err != nil {
		return CustomerReceipt{}, err
	}
	now := time.Now().UTC().Format(time.RFC3339Nano)
	return CustomerReceipt{
		id: uuid.NewString(), customerID: params.CustomerID, receiptNumber: params.ReceiptNumber,
		receiptDate: params.ReceiptDate, currency: params.Currency, exchangeRate: params.ExchangeRate,
		amount: params.Amount, paymentMethod: params.PaymentMethod, paymentReference: params.PaymentReference,
		status: StatusDraft, idempotencyKey: params.IdempotencyKey, allocations: copyAllocations(params.Allocations),
		createdAt: now, updatedAt: now,
	}, nil
}

func validateParams(params NewCustomerReceiptParams) error {
	if params.CustomerID == "" || params.ReceiptNumber <= 0 || params.PaymentMethod == "" ||
		params.IdempotencyKey == "" || len(params.Currency) != 3 {
		return ErrInvalidCustomerReceipt
	}
	if _, err := time.Parse("2006-01-02", params.ReceiptDate); err != nil {
		return ErrInvalidCustomerReceipt
	}
	if valueObject.ValidateDecimal(params.ExchangeRate, 6, true) != nil ||
		valueObject.ValidateDecimal(params.Amount, 4, true) != nil {
		return ErrInvalidCustomerReceipt
	}
	allocated := "0"
	invoices := map[string]struct{}{}
	for _, allocation := range params.Allocations {
		if allocation.SalesInvoiceID == "" || valueObject.ValidateDecimal(allocation.AllocatedAmount, 4, true) != nil {
			return ErrInvalidCustomerReceipt
		}
		if _, exists := invoices[allocation.SalesInvoiceID]; exists {
			return ErrInvalidCustomerReceipt
		}
		invoices[allocation.SalesInvoiceID] = struct{}{}
		allocated, _ = valueObject.AddDecimal(allocated, allocation.AllocatedAmount)
	}
	cmp, err := valueObject.CompareDecimal(allocated, params.Amount)
	if err != nil || cmp > 0 {
		return ErrInvalidCustomerReceipt
	}
	return nil
}

func (r *CustomerReceipt) Post() error {
	if r.status == StatusPosted {
		return nil
	}
	if r.status != StatusDraft {
		return ErrReceiptImmutable
	}
	r.status = StatusPosted
	r.updatedAt = time.Now().UTC().Format(time.RFC3339Nano)
	return nil
}

func (r CustomerReceipt) GetID() string                { return r.id }
func (r CustomerReceipt) GetCustomerID() string        { return r.customerID }
func (r CustomerReceipt) GetReceiptNumber() int64      { return r.receiptNumber }
func (r CustomerReceipt) GetReceiptDate() string       { return r.receiptDate }
func (r CustomerReceipt) GetCurrency() string          { return r.currency }
func (r CustomerReceipt) GetExchangeRate() string      { return r.exchangeRate }
func (r CustomerReceipt) GetAmount() string            { return r.amount }
func (r CustomerReceipt) GetPaymentMethod() string     { return r.paymentMethod }
func (r CustomerReceipt) GetPaymentReference() string  { return r.paymentReference }
func (r CustomerReceipt) GetStatus() Status            { return r.status }
func (r CustomerReceipt) GetIdempotencyKey() string    { return r.idempotencyKey }
func (r CustomerReceipt) GetCreatedAt() string         { return r.createdAt }
func (r CustomerReceipt) GetUpdatedAt() string         { return r.updatedAt }
func (r CustomerReceipt) IsEditable() bool             { return r.status == StatusDraft }
func (r CustomerReceipt) GetAllocations() []Allocation { return copyAllocations(r.allocations) }

func ReconstituteCustomerReceipt(id string, params NewCustomerReceiptParams, status Status, createdAt, updatedAt string) CustomerReceipt {
	return CustomerReceipt{
		id: id, customerID: params.CustomerID, receiptNumber: params.ReceiptNumber, receiptDate: params.ReceiptDate,
		currency: params.Currency, exchangeRate: params.ExchangeRate, amount: params.Amount, paymentMethod: params.PaymentMethod,
		paymentReference: params.PaymentReference, status: status, idempotencyKey: params.IdempotencyKey,
		allocations: copyAllocations(params.Allocations), createdAt: createdAt, updatedAt: updatedAt,
	}
}

func copyAllocations(allocations []Allocation) []Allocation {
	result := make([]Allocation, len(allocations))
	copy(result, allocations)
	return result
}

package purchasing_administration

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

type SupplierPaymentStatus string

const (
	SupplierPaymentDraft  SupplierPaymentStatus = "DRAFT"
	SupplierPaymentPosted SupplierPaymentStatus = "POSTED"
	SupplierPaymentVoided SupplierPaymentStatus = "VOIDED"
)

type SupplierPaymentAllocation struct {
	ID, SupplierPaymentID, SupplierInvoiceID, AllocatedAmount, CreatedAt, UpdatedAt string
}

type SupplierPayment struct {
	ID, SupplierID, PaymentDate, ExchangeRate, Amount, PaymentMethod string
	PaymentReference, IdempotencyKey, CreatedAt, UpdatedAt           string
	PaymentNumber                                                    int64
	Currency                                                         Currency
	Status                                                           SupplierPaymentStatus
	Allocations                                                      []SupplierPaymentAllocation
}

type NewSupplierPaymentParams struct {
	SupplierID, PaymentDate, ExchangeRate, Amount, PaymentMethod string
	PaymentReference, IdempotencyKey                             string
	PaymentNumber                                                int64
	Currency                                                     Currency
	Allocations                                                  []SupplierPaymentAllocation
}

func NewSupplierPayment(params NewSupplierPaymentParams) (SupplierPayment, error) {
	if params.SupplierID == "" || params.PaymentNumber <= 0 {
		return SupplierPayment{}, errors.New("supplier_id and positive payment_number are required")
	}
	if err := validateDate(params.PaymentDate, "payment_date"); err != nil {
		return SupplierPayment{}, err
	}
	if strings.TrimSpace(params.PaymentMethod) == "" || strings.TrimSpace(params.IdempotencyKey) == "" {
		return SupplierPayment{}, errors.New("payment_method and idempotency_key are required")
	}
	if params.Currency == "" {
		params.Currency = CurrencyARS
	}
	if err := validateCurrency(params.Currency); err != nil {
		return SupplierPayment{}, err
	}
	if params.ExchangeRate == "" {
		params.ExchangeRate = "1"
	}
	rate, err := normalizedExchangeRate(params.ExchangeRate)
	if err != nil {
		return SupplierPayment{}, errors.New("exchange_rate must be a positive decimal string")
	}
	amount, err := normalizedMoney(params.Amount, true)
	if err != nil {
		return SupplierPayment{}, errors.New("amount must be a positive decimal string")
	}
	id, now := uuid.NewString(), time.Now().UTC().Format(time.RFC3339Nano)
	allocations := append([]SupplierPaymentAllocation(nil), params.Allocations...)
	seen, allocated := map[string]bool{}, "0"
	for i := range allocations {
		a := &allocations[i]
		if a.SupplierInvoiceID == "" || seen[a.SupplierInvoiceID] {
			return SupplierPayment{}, errors.New("allocation supplier_invoice_id must be present and unique")
		}
		value, err := normalizedMoney(a.AllocatedAmount, true)
		if err != nil {
			return SupplierPayment{}, errors.New("allocated_amount must be a positive decimal string")
		}
		allocated, _ = AddDecimals(allocated, value)
		a.ID, a.SupplierPaymentID, a.AllocatedAmount = uuid.NewString(), id, value
		a.CreatedAt, a.UpdatedAt = now, now
		seen[a.SupplierInvoiceID] = true
	}
	if cmp, _ := CompareDecimals(allocated, amount); cmp > 0 {
		return SupplierPayment{}, errors.New("allocations must not exceed payment amount")
	}
	return SupplierPayment{ID: id, SupplierID: params.SupplierID, PaymentNumber: params.PaymentNumber,
		PaymentDate: params.PaymentDate, Currency: params.Currency, ExchangeRate: rate, Amount: amount,
		PaymentMethod: params.PaymentMethod, PaymentReference: params.PaymentReference, Status: SupplierPaymentDraft,
		IdempotencyKey: params.IdempotencyKey, CreatedAt: now, UpdatedAt: now, Allocations: allocations}, nil
}

func (p *SupplierPayment) Post() error {
	if p.Status == SupplierPaymentPosted {
		return nil
	}
	if p.Status != SupplierPaymentDraft {
		return errors.New("only a draft supplier payment can be posted")
	}
	p.Status, p.UpdatedAt = SupplierPaymentPosted, time.Now().UTC().Format(time.RFC3339Nano)
	return nil
}

func (p *SupplierPayment) Void() error {
	if p.Status == SupplierPaymentVoided {
		return nil
	}
	if p.Status != SupplierPaymentPosted {
		return errors.New("only a posted supplier payment can be voided")
	}
	p.Status, p.UpdatedAt = SupplierPaymentVoided, time.Now().UTC().Format(time.RFC3339Nano)
	return nil
}

type SupplierInvoiceBalance struct {
	SupplierInvoiceID, SupplierID, DocumentType, IssueDate, Currency string
	PointOfSale                                                      int32
	DocumentNumber                                                   int64
	TotalAmount, AllocatedAmount, OutstandingAmount                  string
}

type SupplierOutstandingBalance struct {
	SupplierID, Currency, TotalOutstanding string
	Invoices                               []SupplierInvoiceBalance
}

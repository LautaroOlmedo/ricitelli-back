package sales_invoice

import (
	"errors"
	"time"

	valueObject "ricitelli-back/internal/value-object"

	"github.com/google/uuid"
)

var (
	ErrInvalidSalesInvoice = errors.New("invalid sales invoice")
	ErrInvoiceImmutable    = errors.New("issued sales invoice is immutable")
)

type DocumentType string

const (
	DocumentTypeInvoice DocumentType = "INVOICE"
	DocumentTypeCredit  DocumentType = "CREDIT"
	DocumentTypeDebit   DocumentType = "DEBIT"
)

type Status string

const (
	StatusDraft         Status = "DRAFT"
	StatusIssued        Status = "ISSUED"
	StatusPartiallyPaid Status = "PARTIALLY_PAID"
	StatusPaid          Status = "PAID"
)

type Item struct {
	LineNumber  int32
	ProductID   string
	Description string
	Quantity    string
	UnitPrice   string
	TaxRate     string
	NetAmount   string
	TaxAmount   string
	TotalAmount string
}

type NewSalesInvoiceParams struct {
	CustomerID     string
	SaleOrderID    string
	DocumentType   DocumentType
	PointOfSale    int32
	DocumentNumber int64
	IssueDate      string
	DueDate        string
	Currency       string
	ExchangeRate   string
	Subtotal       string
	TaxTotal       string
	TotalAmount    string
	IdempotencyKey string
	Items          []Item
}

type SalesInvoice struct {
	id             string
	customerID     string
	saleOrderID    string
	documentType   DocumentType
	pointOfSale    int32
	documentNumber int64
	issueDate      string
	dueDate        string
	currency       string
	exchangeRate   string
	subtotal       string
	taxTotal       string
	totalAmount    string
	status         Status
	idempotencyKey string
	items          []Item
	createdAt      string
	updatedAt      string
}

func NewSalesInvoice(params NewSalesInvoiceParams) (SalesInvoice, error) {
	if err := validateParams(params); err != nil {
		return SalesInvoice{}, err
	}
	now := time.Now().UTC().Format(time.RFC3339Nano)
	return SalesInvoice{
		id:             uuid.NewString(),
		customerID:     params.CustomerID,
		saleOrderID:    params.SaleOrderID,
		documentType:   params.DocumentType,
		pointOfSale:    params.PointOfSale,
		documentNumber: params.DocumentNumber,
		issueDate:      params.IssueDate,
		dueDate:        params.DueDate,
		currency:       params.Currency,
		exchangeRate:   params.ExchangeRate,
		subtotal:       params.Subtotal,
		taxTotal:       params.TaxTotal,
		totalAmount:    params.TotalAmount,
		status:         StatusDraft,
		idempotencyKey: params.IdempotencyKey,
		items:          copyItems(params.Items),
		createdAt:      now,
		updatedAt:      now,
	}, nil
}

func validateParams(params NewSalesInvoiceParams) error {
	if params.CustomerID == "" || params.SaleOrderID == "" || params.IdempotencyKey == "" ||
		params.PointOfSale <= 0 || params.DocumentNumber <= 0 || len(params.Items) == 0 {
		return ErrInvalidSalesInvoice
	}
	if params.DocumentType != DocumentTypeInvoice && params.DocumentType != DocumentTypeCredit && params.DocumentType != DocumentTypeDebit {
		return ErrInvalidSalesInvoice
	}
	if !validDateRange(params.IssueDate, params.DueDate) || len(params.Currency) != 3 {
		return ErrInvalidSalesInvoice
	}
	if valueObject.ValidateDecimal(params.ExchangeRate, 6, true) != nil ||
		valueObject.ValidateDecimal(params.Subtotal, 4, false) != nil ||
		valueObject.ValidateDecimal(params.TaxTotal, 4, false) != nil ||
		valueObject.ValidateDecimal(params.TotalAmount, 4, false) != nil {
		return ErrInvalidSalesInvoice
	}

	lineNumbers := map[int32]struct{}{}
	netTotal, taxTotal, itemTotal := "0", "0", "0"
	for _, item := range params.Items {
		if item.LineNumber <= 0 || item.ProductID == "" || item.Description == "" {
			return ErrInvalidSalesInvoice
		}
		if _, exists := lineNumbers[item.LineNumber]; exists {
			return ErrInvalidSalesInvoice
		}
		lineNumbers[item.LineNumber] = struct{}{}
		if valueObject.ValidateDecimal(item.Quantity, 4, true) != nil ||
			valueObject.ValidateDecimal(item.UnitPrice, 4, false) != nil ||
			valueObject.ValidateDecimal(item.TaxRate, 4, false) != nil ||
			valueObject.ValidateDecimal(item.NetAmount, 4, false) != nil ||
			valueObject.ValidateDecimal(item.TaxAmount, 4, false) != nil ||
			valueObject.ValidateDecimal(item.TotalAmount, 4, false) != nil {
			return ErrInvalidSalesInvoice
		}
		var err error
		netTotal, err = valueObject.AddDecimal(netTotal, item.NetAmount)
		if err != nil {
			return ErrInvalidSalesInvoice
		}
		taxTotal, _ = valueObject.AddDecimal(taxTotal, item.TaxAmount)
		itemTotal, _ = valueObject.AddDecimal(itemTotal, item.TotalAmount)
	}
	computedTotal, _ := valueObject.AddDecimal(params.Subtotal, params.TaxTotal)
	if !decimalsEqual(netTotal, params.Subtotal) || !decimalsEqual(taxTotal, params.TaxTotal) ||
		!decimalsEqual(itemTotal, params.TotalAmount) || !decimalsEqual(computedTotal, params.TotalAmount) {
		return ErrInvalidSalesInvoice
	}
	return nil
}

func (s *SalesInvoice) Issue() error {
	if s.status == StatusIssued {
		return nil
	}
	if s.status != StatusDraft {
		return ErrInvoiceImmutable
	}
	s.status = StatusIssued
	s.updatedAt = time.Now().UTC().Format(time.RFC3339Nano)
	return nil
}

func (s *SalesInvoice) UpdatePaymentStatus(paid bool) {
	if paid {
		s.status = StatusPaid
	} else {
		s.status = StatusPartiallyPaid
	}
	s.updatedAt = time.Now().UTC().Format(time.RFC3339Nano)
}

func (s SalesInvoice) GetID() string                 { return s.id }
func (s SalesInvoice) GetCustomerID() string         { return s.customerID }
func (s SalesInvoice) GetSaleOrderID() string        { return s.saleOrderID }
func (s SalesInvoice) GetDocumentType() DocumentType { return s.documentType }
func (s SalesInvoice) GetPointOfSale() int32         { return s.pointOfSale }
func (s SalesInvoice) GetDocumentNumber() int64      { return s.documentNumber }
func (s SalesInvoice) GetIssueDate() string          { return s.issueDate }
func (s SalesInvoice) GetDueDate() string            { return s.dueDate }
func (s SalesInvoice) GetCurrency() string           { return s.currency }
func (s SalesInvoice) GetExchangeRate() string       { return s.exchangeRate }
func (s SalesInvoice) GetSubtotal() string           { return s.subtotal }
func (s SalesInvoice) GetTaxTotal() string           { return s.taxTotal }
func (s SalesInvoice) GetTotalAmount() string        { return s.totalAmount }
func (s SalesInvoice) GetStatus() Status             { return s.status }
func (s SalesInvoice) GetIdempotencyKey() string     { return s.idempotencyKey }
func (s SalesInvoice) GetCreatedAt() string          { return s.createdAt }
func (s SalesInvoice) GetUpdatedAt() string          { return s.updatedAt }
func (s SalesInvoice) IsEditable() bool              { return s.status == StatusDraft }
func (s SalesInvoice) GetItems() []Item              { return copyItems(s.items) }

func ReconstituteSalesInvoice(id string, params NewSalesInvoiceParams, status Status, createdAt, updatedAt string) SalesInvoice {
	return SalesInvoice{
		id: id, customerID: params.CustomerID, saleOrderID: params.SaleOrderID, documentType: params.DocumentType,
		pointOfSale: params.PointOfSale, documentNumber: params.DocumentNumber, issueDate: params.IssueDate,
		dueDate: params.DueDate, currency: params.Currency, exchangeRate: params.ExchangeRate, subtotal: params.Subtotal,
		taxTotal: params.TaxTotal, totalAmount: params.TotalAmount, status: status, idempotencyKey: params.IdempotencyKey,
		items: copyItems(params.Items), createdAt: createdAt, updatedAt: updatedAt,
	}
}

func copyItems(items []Item) []Item {
	result := make([]Item, len(items))
	copy(result, items)
	return result
}

func validDateRange(from, to string) bool {
	fromDate, err := time.Parse("2006-01-02", from)
	if err != nil {
		return false
	}
	if to == "" {
		return true
	}
	toDate, err := time.Parse("2006-01-02", to)
	return err == nil && !toDate.Before(fromDate)
}

func decimalsEqual(left, right string) bool {
	cmp, err := valueObject.CompareDecimal(left, right)
	return err == nil && cmp == 0
}

package purchasing_administration

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

type SupplierInvoiceDocumentType string

const (
	SupplierInvoiceTypeInvoice    SupplierInvoiceDocumentType = "INVOICE"
	SupplierInvoiceTypeCreditNote SupplierInvoiceDocumentType = "CREDIT_NOTE"
	SupplierInvoiceTypeDebitNote  SupplierInvoiceDocumentType = "DEBIT_NOTE"
)

type SupplierInvoiceStatus string

const (
	SupplierInvoiceDraft         SupplierInvoiceStatus = "DRAFT"
	SupplierInvoiceIssued        SupplierInvoiceStatus = "ISSUED"
	SupplierInvoicePartiallyPaid SupplierInvoiceStatus = "PARTIALLY_PAID"
	SupplierInvoicePaid          SupplierInvoiceStatus = "PAID"
	SupplierInvoiceCancelled     SupplierInvoiceStatus = "CANCELLED"
)

type SupplierInvoiceItem struct {
	ID, SupplierInvoiceID, SupplierQuoteItemID, DrySupplyID, Description, Quantity, Unit string
	UnitPrice, TaxRate, NetAmount, TaxAmount, TotalAmount, CreatedAt, UpdatedAt          string
	LineNumber                                                                           int32
}

type SupplierInvoice struct {
	ID, SupplierID, SupplierQuoteID, IssueDate, DueDate, ExchangeRate     string
	Subtotal, TaxTotal, TotalAmount, IdempotencyKey, CreatedAt, UpdatedAt string
	DocumentType                                                          SupplierInvoiceDocumentType
	PointOfSale                                                           int32
	DocumentNumber                                                        int64
	Currency                                                              Currency
	Status                                                                SupplierInvoiceStatus
	Items                                                                 []SupplierInvoiceItem
}

type NewSupplierInvoiceParams struct {
	SupplierID, SupplierQuoteID, IssueDate, DueDate, ExchangeRate string
	Subtotal, TaxTotal, TotalAmount, IdempotencyKey               string
	DocumentType                                                  SupplierInvoiceDocumentType
	PointOfSale                                                   int32
	DocumentNumber                                                int64
	Currency                                                      Currency
	Items                                                         []SupplierInvoiceItem
}

func NewSupplierInvoice(params NewSupplierInvoiceParams) (SupplierInvoice, error) {
	if params.SupplierID == "" {
		return SupplierInvoice{}, errors.New("supplier_id is required")
	}
	if params.DocumentType != SupplierInvoiceTypeInvoice && params.DocumentType != SupplierInvoiceTypeCreditNote && params.DocumentType != SupplierInvoiceTypeDebitNote {
		return SupplierInvoice{}, errors.New("invalid supplier invoice document_type")
	}
	if params.PointOfSale <= 0 || params.DocumentNumber <= 0 {
		return SupplierInvoice{}, errors.New("point_of_sale and document_number must be greater than zero")
	}
	if err := validateDate(params.IssueDate, "issue_date"); err != nil {
		return SupplierInvoice{}, err
	}
	if err := validateOptionalDate(params.DueDate, "due_date"); err != nil {
		return SupplierInvoice{}, err
	}
	if err := validateDateOrder(params.IssueDate, params.DueDate, "issue_date", "due_date"); err != nil {
		return SupplierInvoice{}, err
	}
	if params.Currency == "" {
		params.Currency = CurrencyARS
	}
	if err := validateCurrency(params.Currency); err != nil {
		return SupplierInvoice{}, err
	}
	if params.ExchangeRate == "" {
		params.ExchangeRate = "1"
	}
	exchangeRate, err := normalizedExchangeRate(params.ExchangeRate)
	if err != nil {
		return SupplierInvoice{}, errors.New("exchange_rate must be a positive decimal string")
	}
	subtotal, tax, total, items, err := validateInvoiceAmounts(params.Subtotal, params.TaxTotal, params.TotalAmount, params.Items)
	if err != nil {
		return SupplierInvoice{}, err
	}
	if strings.TrimSpace(params.IdempotencyKey) == "" {
		return SupplierInvoice{}, errors.New("idempotency_key is required")
	}
	id, now := uuid.NewString(), time.Now().UTC().Format(time.RFC3339Nano)
	for i := range items {
		items[i].ID, items[i].SupplierInvoiceID, items[i].CreatedAt, items[i].UpdatedAt = uuid.NewString(), id, now, now
	}
	return SupplierInvoice{ID: id, SupplierID: params.SupplierID, SupplierQuoteID: params.SupplierQuoteID,
		DocumentType: params.DocumentType, PointOfSale: params.PointOfSale, DocumentNumber: params.DocumentNumber,
		IssueDate: params.IssueDate, DueDate: params.DueDate, Currency: params.Currency, ExchangeRate: exchangeRate,
		Subtotal: subtotal, TaxTotal: tax, TotalAmount: total, Status: SupplierInvoiceDraft,
		IdempotencyKey: params.IdempotencyKey, CreatedAt: now, UpdatedAt: now, Items: items}, nil
}

func validateInvoiceAmounts(subtotalValue, taxValue, totalValue string, input []SupplierInvoiceItem) (string, string, string, []SupplierInvoiceItem, error) {
	quoteItems := make([]SupplierQuoteItem, len(input))
	for i, item := range input {
		quoteItems[i] = SupplierQuoteItem{
			LineNumber: item.LineNumber, DrySupplyID: item.DrySupplyID, Description: item.Description,
			Quantity: item.Quantity, Unit: item.Unit, UnitPrice: item.UnitPrice, TaxRate: item.TaxRate,
			NetAmount: item.NetAmount, TaxAmount: item.TaxAmount, TotalAmount: item.TotalAmount,
		}
	}
	subtotal, tax, total, validated, err := validateDocumentAmounts(subtotalValue, taxValue, totalValue, quoteItems)
	if err != nil {
		return "", "", "", nil, err
	}
	items := append([]SupplierInvoiceItem(nil), input...)
	for i := range items {
		items[i].Quantity, items[i].UnitPrice, items[i].TaxRate = validated[i].Quantity, validated[i].UnitPrice, validated[i].TaxRate
		items[i].NetAmount, items[i].TaxAmount, items[i].TotalAmount = validated[i].NetAmount, validated[i].TaxAmount, validated[i].TotalAmount
	}
	return subtotal, tax, total, items, nil
}

func (i *SupplierInvoice) Issue() error {
	if i.Status == SupplierInvoiceIssued {
		return nil
	}
	if i.Status != SupplierInvoiceDraft {
		return errors.New("only a draft supplier invoice can be issued")
	}
	i.Status, i.UpdatedAt = SupplierInvoiceIssued, time.Now().UTC().Format(time.RFC3339Nano)
	return nil
}

// Void cancels an invoice while preserving it for audit and VAT reporting.
func (i *SupplierInvoice) Void() error {
	if i.Status == SupplierInvoiceCancelled {
		return nil
	}
	if i.Status == SupplierInvoicePartiallyPaid || i.Status == SupplierInvoicePaid {
		return errors.New("supplier invoice with posted payments cannot be voided")
	}
	i.Status, i.UpdatedAt = SupplierInvoiceCancelled, time.Now().UTC().Format(time.RFC3339Nano)
	return nil
}

func (i *SupplierInvoice) SetPaymentStatus(status SupplierInvoiceStatus) error {
	if i.Status != SupplierInvoiceIssued && i.Status != SupplierInvoicePartiallyPaid {
		return errors.New("supplier invoice must be issued before recording payments")
	}
	if status != SupplierInvoicePartiallyPaid && status != SupplierInvoicePaid {
		return errors.New("invalid supplier invoice payment status")
	}
	i.Status, i.UpdatedAt = status, time.Now().UTC().Format(time.RFC3339Nano)
	return nil
}

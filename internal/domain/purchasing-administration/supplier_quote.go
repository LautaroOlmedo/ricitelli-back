package purchasing_administration

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

type SupplierQuoteStatus string

const (
	SupplierQuoteDraft     SupplierQuoteStatus = "DRAFT"
	SupplierQuoteReceived  SupplierQuoteStatus = "RECEIVED"
	SupplierQuoteAccepted  SupplierQuoteStatus = "ACCEPTED"
	SupplierQuoteRejected  SupplierQuoteStatus = "REJECTED"
	SupplierQuoteExpired   SupplierQuoteStatus = "EXPIRED"
	SupplierQuoteCancelled SupplierQuoteStatus = "CANCELLED"
)

type SupplierQuoteItem struct {
	ID, SupplierQuoteID, PurchaseNeedItemID, DrySupplyID, Description, Quantity, Unit string
	UnitPrice, TaxRate, NetAmount, TaxAmount, TotalAmount, CreatedAt, UpdatedAt       string
	LineNumber                                                                        int32
}

type SupplierQuote struct {
	ID, SupplierID, PurchaseNeedID, QuoteNumber, QuoteDate, ValidUntil, ExchangeRate string
	Subtotal, TaxTotal, TotalAmount, IdempotencyKey, CreatedAt, UpdatedAt            string
	Currency                                                                         Currency
	Status                                                                           SupplierQuoteStatus
	Items                                                                            []SupplierQuoteItem
}

type NewSupplierQuoteParams struct {
	SupplierID, PurchaseNeedID, QuoteNumber, QuoteDate, ValidUntil, ExchangeRate string
	Subtotal, TaxTotal, TotalAmount, IdempotencyKey                              string
	Currency                                                                     Currency
	Items                                                                        []SupplierQuoteItem
}

func NewSupplierQuote(params NewSupplierQuoteParams) (SupplierQuote, error) {
	if params.SupplierID == "" || strings.TrimSpace(params.QuoteNumber) == "" {
		return SupplierQuote{}, errors.New("supplier_id and quote_number are required")
	}
	if err := validateDate(params.QuoteDate, "quote_date"); err != nil {
		return SupplierQuote{}, err
	}
	if err := validateOptionalDate(params.ValidUntil, "valid_until"); err != nil {
		return SupplierQuote{}, err
	}
	if err := validateDateOrder(params.QuoteDate, params.ValidUntil, "quote_date", "valid_until"); err != nil {
		return SupplierQuote{}, err
	}
	if params.Currency == "" {
		params.Currency = CurrencyARS
	}
	if err := validateCurrency(params.Currency); err != nil {
		return SupplierQuote{}, err
	}
	if params.ExchangeRate == "" {
		params.ExchangeRate = "1"
	}
	exchangeRate, err := normalizedExchangeRate(params.ExchangeRate)
	if err != nil {
		return SupplierQuote{}, errors.New("exchange_rate must be a positive decimal string")
	}
	subtotal, tax, total, items, err := validateDocumentAmounts(params.Subtotal, params.TaxTotal, params.TotalAmount, params.Items)
	if err != nil {
		return SupplierQuote{}, err
	}
	if strings.TrimSpace(params.IdempotencyKey) == "" {
		return SupplierQuote{}, errors.New("idempotency_key is required")
	}
	id, now := uuid.NewString(), time.Now().UTC().Format(time.RFC3339Nano)
	for i := range items {
		items[i].ID, items[i].SupplierQuoteID, items[i].CreatedAt, items[i].UpdatedAt = uuid.NewString(), id, now, now
	}
	return SupplierQuote{ID: id, SupplierID: params.SupplierID, PurchaseNeedID: params.PurchaseNeedID,
		QuoteNumber: params.QuoteNumber, QuoteDate: params.QuoteDate, ValidUntil: params.ValidUntil,
		Currency: params.Currency, ExchangeRate: exchangeRate, Subtotal: subtotal, TaxTotal: tax,
		TotalAmount: total, Status: SupplierQuoteDraft, IdempotencyKey: params.IdempotencyKey,
		CreatedAt: now, UpdatedAt: now, Items: items}, nil
}

func validateDocumentAmounts(subtotalValue, taxValue, totalValue string, input []SupplierQuoteItem) (string, string, string, []SupplierQuoteItem, error) {
	if len(input) == 0 {
		return "", "", "", nil, errors.New("document must have at least one item")
	}
	subtotal, err := normalizedMoney(subtotalValue, false)
	if err != nil {
		return "", "", "", nil, errors.New("subtotal must be a non-negative decimal string")
	}
	tax, err := normalizedMoney(taxValue, false)
	if err != nil {
		return "", "", "", nil, errors.New("tax_total must be a non-negative decimal string")
	}
	total, err := normalizedMoney(totalValue, false)
	if err != nil {
		return "", "", "", nil, errors.New("total_amount must be a non-negative decimal string")
	}
	sum, _ := AddDecimals(subtotal, tax)
	if cmp, _ := CompareDecimals(sum, total); cmp != 0 {
		return "", "", "", nil, errors.New("total_amount must equal subtotal plus tax_total")
	}
	items := append([]SupplierQuoteItem(nil), input...)
	seen := map[int32]bool{}
	itemSubtotal, itemTax, itemTotal := "0", "0", "0"
	for i := range items {
		item := &items[i]
		if item.LineNumber <= 0 || seen[item.LineNumber] {
			return "", "", "", nil, errors.New("item line_number must be positive and unique")
		}
		if strings.TrimSpace(item.Description) == "" || strings.TrimSpace(item.Unit) == "" {
			return "", "", "", nil, errors.New("item description and unit are required")
		}
		for label, value := range map[string]string{
			"quantity": item.Quantity, "unit_price": item.UnitPrice, "tax_rate": item.TaxRate,
			"net_amount": item.NetAmount, "tax_amount": item.TaxAmount, "total_amount": item.TotalAmount,
		} {
			positive := label == "quantity"
			var n string
			var e error
			if label == "tax_rate" {
				n, e = normalizedTaxRate(value)
			} else {
				n, e = normalizedMoney(value, positive)
			}
			if e != nil {
				return "", "", "", nil, errors.New("item " + label + " is not a valid decimal string")
			}
			switch label {
			case "quantity":
				item.Quantity = n
			case "unit_price":
				item.UnitPrice = n
			case "tax_rate":
				item.TaxRate = n
			case "net_amount":
				item.NetAmount = n
			case "tax_amount":
				item.TaxAmount = n
			case "total_amount":
				item.TotalAmount = n
			}
		}
		lineTotal, _ := AddDecimals(item.NetAmount, item.TaxAmount)
		if cmp, _ := CompareDecimals(lineTotal, item.TotalAmount); cmp != 0 {
			return "", "", "", nil, errors.New("item total_amount must equal net_amount plus tax_amount")
		}
		itemSubtotal, _ = AddDecimals(itemSubtotal, item.NetAmount)
		itemTax, _ = AddDecimals(itemTax, item.TaxAmount)
		itemTotal, _ = AddDecimals(itemTotal, item.TotalAmount)
		seen[item.LineNumber] = true
	}
	if cmp, _ := CompareDecimals(itemSubtotal, subtotal); cmp != 0 {
		return "", "", "", nil, errors.New("subtotal must equal the sum of item net_amount values")
	}
	if cmp, _ := CompareDecimals(itemTax, tax); cmp != 0 {
		return "", "", "", nil, errors.New("tax_total must equal the sum of item tax_amount values")
	}
	if cmp, _ := CompareDecimals(itemTotal, total); cmp != 0 {
		return "", "", "", nil, errors.New("total_amount must equal the sum of item total_amount values")
	}
	return subtotal, tax, total, items, nil
}

var quoteTransitions = map[SupplierQuoteStatus]map[SupplierQuoteStatus]bool{
	SupplierQuoteDraft:    {SupplierQuoteReceived: true, SupplierQuoteCancelled: true},
	SupplierQuoteReceived: {SupplierQuoteAccepted: true, SupplierQuoteRejected: true, SupplierQuoteExpired: true, SupplierQuoteCancelled: true},
}

func (q *SupplierQuote) UpdateStatus(status SupplierQuoteStatus) error {
	if !quoteTransitions[q.Status][status] {
		return errors.New("invalid supplier quote status transition")
	}
	q.Status, q.UpdatedAt = status, time.Now().UTC().Format(time.RFC3339Nano)
	return nil
}

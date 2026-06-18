package sale_order

import (
	"context"
	"errors"
	"testing"

	"ricitelli-back/internal/domain/remittance"
	sale_order "ricitelli-back/internal/domain/sale-order"
	sales_invoice "ricitelli-back/internal/domain/sales-invoice"
	sales_administration "ricitelli-back/internal/service/sales-administration"
	valueObject "ricitelli-back/internal/value-object"
)

func TestGetSaleOrderByIDDerivesAdministrativeSummaryWithoutChangingLegacyStatus(t *testing.T) {
	order := sale_order.ReconstitueSaleOrder(
		"order-1", "customer-1", sale_order.StatusInvoiced,
		[]valueObject.SaleOrderItem{{ProductID: "p1", Quantity: 10}, {ProductID: "p2", Quantity: 5}},
		sale_order.CurrencyARS, sale_order.MarketDomestic, "AR", sale_order.SaleTypeRegular, "2026-06-09T00:00:00Z", true,
	)
	storage := &summaryStorage{
		order: order,
		invoices: []sales_invoice.SalesInvoice{
			sales_invoice.ReconstituteSalesInvoice("invoice-issued", sales_invoice.NewSalesInvoiceParams{
				CustomerID: "customer-1", SaleOrderID: "order-1", DocumentType: sales_invoice.DocumentTypeInvoice,
				PointOfSale: 1, DocumentNumber: 2, Items: []sales_invoice.Item{{Quantity: "4"}},
			}, sales_invoice.StatusIssued, "", ""),
			sales_invoice.ReconstituteSalesInvoice("invoice-draft", sales_invoice.NewSalesInvoiceParams{
				CustomerID: "customer-1", SaleOrderID: "order-1", DocumentType: sales_invoice.DocumentTypeInvoice,
				PointOfSale: 1, DocumentNumber: 3, Items: []sales_invoice.Item{{Quantity: "2"}},
			}, sales_invoice.StatusDraft, "", ""),
		},
		remittances: []remittance.Remittance{
			remittance.ReconstituteRemittance("remittance-confirmed", remittance.NewRemittanceParams{
				CustomerID: "customer-1", SaleOrderID: "order-1", PointOfSale: 2, DocumentNumber: 4,
				Items: []remittance.Item{{Quantity: "3"}},
			}, remittance.StatusConfirmed, "", ""),
			remittance.ReconstituteRemittance("remittance-draft", remittance.NewRemittanceParams{
				CustomerID: "customer-1", SaleOrderID: "order-1", PointOfSale: 2, DocumentNumber: 5,
				Items: []remittance.Item{{Quantity: "1"}},
			}, remittance.StatusDraft, "", ""),
		},
	}

	got, err := NewSaleOrderService(storage).GetSaleOrderByID(context.Background(), "order-1")
	if err != nil {
		t.Fatalf("GetSaleOrderByID() error = %v", err)
	}
	if got.GetStatus() != sale_order.StatusInvoiced {
		t.Fatalf("status = %q, want legacy status %q unchanged", got.GetStatus(), sale_order.StatusInvoiced)
	}
	summary := got.GetAdministrativeSummary()
	if summary.TotalOrderedQuantity != "15" || summary.InvoicedQuantity != "4" || summary.RemittedQuantity != "3" ||
		summary.PendingInvoiceQuantity != "11" || summary.PendingRemittanceQuantity != "12" {
		t.Fatalf("summary = %+v", summary)
	}
	if len(summary.LinkedInvoices) != 2 || len(summary.LinkedRemittances) != 2 {
		t.Fatalf("linked documents = %d invoices, %d remittances", len(summary.LinkedInvoices), len(summary.LinkedRemittances))
	}
	if summary.LinkedInvoices[0].DocumentNumber != "00001-00000002" || summary.LinkedRemittances[0].DocumentNumber != "00002-00000004" {
		t.Fatalf("document identifiers = %+v / %+v", summary.LinkedInvoices, summary.LinkedRemittances)
	}
}

type summaryStorage struct {
	order       sale_order.SaleOrder
	invoices    []sales_invoice.SalesInvoice
	remittances []remittance.Remittance
}

func (s *summaryStorage) CreateSaleOrder(context.Context, sale_order.NewSaleOrderParams) (sale_order.SaleOrder, error) {
	return sale_order.SaleOrder{}, errors.New("unexpected CreateSaleOrder")
}
func (s *summaryStorage) GetSaleOrderByID(context.Context, string) (*sale_order.SaleOrder, error) {
	order := s.order
	return &order, nil
}
func (s *summaryStorage) GetSaleOrders(context.Context) ([]sale_order.SaleOrder, error) {
	return []sale_order.SaleOrder{s.order}, nil
}
func (s *summaryStorage) GetSaleOrdersByDateRange(context.Context, string, string) ([]sale_order.SaleOrder, error) {
	return []sale_order.SaleOrder{s.order}, nil
}
func (s *summaryStorage) UpdateSaleOrderStatus(context.Context, string, sale_order.Status) (*sale_order.SaleOrder, error) {
	return nil, errors.New("unexpected UpdateSaleOrderStatus")
}
func (s *summaryStorage) ListSalesInvoices(context.Context, sales_administration.SalesInvoiceFilter) ([]sales_invoice.SalesInvoice, error) {
	return s.invoices, nil
}
func (s *summaryStorage) ListRemittances(context.Context, sales_administration.RemittanceFilter) ([]remittance.Remittance, error) {
	return s.remittances, nil
}

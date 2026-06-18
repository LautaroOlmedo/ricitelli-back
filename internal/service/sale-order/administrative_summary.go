package sale_order

import (
	"context"
	"fmt"

	"ricitelli-back/internal/domain/remittance"
	sale_order "ricitelli-back/internal/domain/sale-order"
	sales_invoice "ricitelli-back/internal/domain/sales-invoice"
	sales_administration "ricitelli-back/internal/service/sales-administration"
	valueObject "ricitelli-back/internal/value-object"
)

func (s *Service) attachAdministrativeSummaries(ctx context.Context, orders []sale_order.SaleOrder) error {
	for i := range orders {
		if err := s.attachAdministrativeSummary(ctx, &orders[i]); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) attachAdministrativeSummary(ctx context.Context, order *sale_order.SaleOrder) error {
	invoices, err := s.Storage.ListSalesInvoices(ctx, sales_administration.SalesInvoiceFilter{SaleOrderID: order.GetID()})
	if err != nil {
		return fmt.Errorf("list linked sales invoices: %w", err)
	}
	remittances, err := s.Storage.ListRemittances(ctx, sales_administration.RemittanceFilter{SaleOrderID: order.GetID()})
	if err != nil {
		return fmt.Errorf("list linked remittances: %w", err)
	}

	summary := sale_order.AdministrativeSummary{
		TotalOrderedQuantity:      "0",
		InvoicedQuantity:          "0",
		RemittedQuantity:          "0",
		PendingInvoiceQuantity:    "0",
		PendingRemittanceQuantity: "0",
		LinkedInvoices:            make([]sale_order.AdministrativeDocumentReference, 0, len(invoices)),
		LinkedRemittances:         make([]sale_order.AdministrativeDocumentReference, 0, len(remittances)),
	}
	for _, item := range order.GetItems() {
		summary.TotalOrderedQuantity, _ = valueObject.AddDecimal(summary.TotalOrderedQuantity, fmt.Sprint(item.Quantity))
	}
	for _, invoice := range invoices {
		summary.LinkedInvoices = append(summary.LinkedInvoices, sale_order.AdministrativeDocumentReference{
			ID: invoice.GetID(), DocumentNumber: documentNumber(invoice.GetPointOfSale(), invoice.GetDocumentNumber()), Status: string(invoice.GetStatus()),
		})
		if invoice.GetStatus() != sales_invoice.StatusIssued || invoice.GetDocumentType() != sales_invoice.DocumentTypeInvoice {
			continue
		}
		for _, item := range invoice.GetItems() {
			summary.InvoicedQuantity, err = valueObject.AddDecimal(summary.InvoicedQuantity, item.Quantity)
			if err != nil {
				return fmt.Errorf("sum issued invoice quantity: %w", err)
			}
		}
	}
	for _, document := range remittances {
		summary.LinkedRemittances = append(summary.LinkedRemittances, sale_order.AdministrativeDocumentReference{
			ID: document.GetID(), DocumentNumber: documentNumber(document.GetPointOfSale(), document.GetDocumentNumber()), Status: string(document.GetStatus()),
		})
		if document.GetStatus() != remittance.StatusConfirmed {
			continue
		}
		for _, item := range document.GetItems() {
			summary.RemittedQuantity, err = valueObject.AddDecimal(summary.RemittedQuantity, item.Quantity)
			if err != nil {
				return fmt.Errorf("sum confirmed remittance quantity: %w", err)
			}
		}
	}
	summary.PendingInvoiceQuantity = nonNegativeDifference(summary.TotalOrderedQuantity, summary.InvoicedQuantity)
	summary.PendingRemittanceQuantity = nonNegativeDifference(summary.TotalOrderedQuantity, summary.RemittedQuantity)
	order.SetAdministrativeSummary(summary)
	return nil
}

func nonNegativeDifference(total, completed string) string {
	result, err := valueObject.SubtractDecimal(total, completed)
	if err != nil {
		return "0"
	}
	if cmp, err := valueObject.CompareDecimal(result, "0"); err != nil || cmp < 0 {
		return "0"
	}
	return result
}

func documentNumber(pointOfSale int32, number int64) string {
	return fmt.Sprintf("%05d-%08d", pointOfSale, number)
}

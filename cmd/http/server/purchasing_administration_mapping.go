package server

import (
	purchasingpb "ricitelli-back/cmd/http/gen/purchasing_administration"
	domain "ricitelli-back/internal/domain/purchasing-administration"
)

func toProtoSupplier(v *domain.Supplier) *purchasingpb.Supplier {
	if v == nil {
		return nil
	}
	return &purchasingpb.Supplier{Id: v.ID, TaxId: v.TaxID, SocialReason: v.SocialReason, TradeName: v.TradeName,
		Email: v.Email, Phone: v.Phone, Address: v.Address, Active: v.Active, IdempotencyKey: v.IdempotencyKey,
		CreatedAt: v.CreatedAt, UpdatedAt: v.UpdatedAt}
}

func purchaseNeedItemsFromProto(items []*purchasingpb.PurchaseNeedItem) []domain.PurchaseNeedItem {
	result := make([]domain.PurchaseNeedItem, len(items))
	for i, item := range items {
		result[i] = domain.PurchaseNeedItem{LineNumber: item.LineNumber, DrySupplyID: item.DrySupplyId,
			Description: item.Description, Quantity: item.Quantity, Unit: item.Unit}
	}
	return result
}

func toProtoPurchaseNeed(v *domain.PurchaseNeed) *purchasingpb.PurchaseNeed {
	if v == nil {
		return nil
	}
	items := make([]*purchasingpb.PurchaseNeedItem, len(v.Items))
	for i, item := range v.Items {
		items[i] = &purchasingpb.PurchaseNeedItem{Id: item.ID, PurchaseNeedId: item.PurchaseNeedID,
			LineNumber: item.LineNumber, DrySupplyId: item.DrySupplyID, Description: item.Description,
			Quantity: item.Quantity, Unit: item.Unit}
	}
	return &purchasingpb.PurchaseNeed{Id: v.ID, NeedNumber: v.NeedNumber, RequestedDate: v.RequestedDate,
		RequiredByDate: v.RequiredByDate, Status: string(v.Status), Notes: v.Notes, IdempotencyKey: v.IdempotencyKey,
		CreatedAt: v.CreatedAt, UpdatedAt: v.UpdatedAt, Items: items}
}

func quoteItemsFromProto(items []*purchasingpb.SupplierQuoteItem) []domain.SupplierQuoteItem {
	result := make([]domain.SupplierQuoteItem, len(items))
	for i, item := range items {
		result[i] = domain.SupplierQuoteItem{PurchaseNeedItemID: item.PurchaseNeedItemId, LineNumber: item.LineNumber,
			DrySupplyID: item.DrySupplyId, Description: item.Description, Quantity: item.Quantity, Unit: item.Unit,
			UnitPrice: item.UnitPrice, TaxRate: item.TaxRate, NetAmount: item.NetAmount, TaxAmount: item.TaxAmount,
			TotalAmount: item.TotalAmount}
	}
	return result
}

func toProtoSupplierQuote(v *domain.SupplierQuote) *purchasingpb.SupplierQuote {
	if v == nil {
		return nil
	}
	items := make([]*purchasingpb.SupplierQuoteItem, len(v.Items))
	for i, item := range v.Items {
		items[i] = &purchasingpb.SupplierQuoteItem{Id: item.ID, SupplierQuoteId: item.SupplierQuoteID,
			PurchaseNeedItemId: item.PurchaseNeedItemID, LineNumber: item.LineNumber, DrySupplyId: item.DrySupplyID,
			Description: item.Description, Quantity: item.Quantity, Unit: item.Unit, UnitPrice: item.UnitPrice,
			TaxRate: item.TaxRate, NetAmount: item.NetAmount, TaxAmount: item.TaxAmount, TotalAmount: item.TotalAmount}
	}
	return &purchasingpb.SupplierQuote{Id: v.ID, SupplierId: v.SupplierID, PurchaseNeedId: v.PurchaseNeedID,
		QuoteNumber: v.QuoteNumber, QuoteDate: v.QuoteDate, ValidUntil: v.ValidUntil, Currency: string(v.Currency),
		ExchangeRate: v.ExchangeRate, Subtotal: v.Subtotal, TaxTotal: v.TaxTotal, TotalAmount: v.TotalAmount,
		Status: string(v.Status), IdempotencyKey: v.IdempotencyKey, CreatedAt: v.CreatedAt, UpdatedAt: v.UpdatedAt, Items: items}
}

func invoiceItemsFromProto(items []*purchasingpb.SupplierInvoiceItem) []domain.SupplierInvoiceItem {
	result := make([]domain.SupplierInvoiceItem, len(items))
	for i, item := range items {
		result[i] = domain.SupplierInvoiceItem{SupplierQuoteItemID: item.SupplierQuoteItemId, LineNumber: item.LineNumber,
			DrySupplyID: item.DrySupplyId, Description: item.Description, Quantity: item.Quantity, Unit: item.Unit,
			UnitPrice: item.UnitPrice, TaxRate: item.TaxRate, NetAmount: item.NetAmount, TaxAmount: item.TaxAmount,
			TotalAmount: item.TotalAmount}
	}
	return result
}

func toProtoSupplierInvoice(v *domain.SupplierInvoice) *purchasingpb.SupplierInvoice {
	if v == nil {
		return nil
	}
	items := make([]*purchasingpb.SupplierInvoiceItem, len(v.Items))
	for i, item := range v.Items {
		items[i] = &purchasingpb.SupplierInvoiceItem{Id: item.ID, SupplierInvoiceId: item.SupplierInvoiceID,
			SupplierQuoteItemId: item.SupplierQuoteItemID, LineNumber: item.LineNumber, DrySupplyId: item.DrySupplyID,
			Description: item.Description, Quantity: item.Quantity, Unit: item.Unit, UnitPrice: item.UnitPrice,
			TaxRate: item.TaxRate, NetAmount: item.NetAmount, TaxAmount: item.TaxAmount, TotalAmount: item.TotalAmount}
	}
	return &purchasingpb.SupplierInvoice{Id: v.ID, SupplierId: v.SupplierID, SupplierQuoteId: v.SupplierQuoteID,
		DocumentType: string(v.DocumentType), PointOfSale: v.PointOfSale, DocumentNumber: v.DocumentNumber,
		IssueDate: v.IssueDate, DueDate: v.DueDate, Currency: string(v.Currency), ExchangeRate: v.ExchangeRate,
		Subtotal: v.Subtotal, TaxTotal: v.TaxTotal, TotalAmount: v.TotalAmount, Status: string(v.Status),
		IdempotencyKey: v.IdempotencyKey, CreatedAt: v.CreatedAt, UpdatedAt: v.UpdatedAt, Items: items}
}

func paymentAllocationsFromProto(items []*purchasingpb.SupplierPaymentAllocation) []domain.SupplierPaymentAllocation {
	result := make([]domain.SupplierPaymentAllocation, len(items))
	for i, item := range items {
		result[i] = domain.SupplierPaymentAllocation{SupplierInvoiceID: item.SupplierInvoiceId, AllocatedAmount: item.AllocatedAmount}
	}
	return result
}

func toProtoSupplierPayment(v *domain.SupplierPayment) *purchasingpb.SupplierPayment {
	if v == nil {
		return nil
	}
	items := make([]*purchasingpb.SupplierPaymentAllocation, len(v.Allocations))
	for i, item := range v.Allocations {
		items[i] = &purchasingpb.SupplierPaymentAllocation{Id: item.ID, SupplierPaymentId: item.SupplierPaymentID,
			SupplierInvoiceId: item.SupplierInvoiceID, AllocatedAmount: item.AllocatedAmount}
	}
	return &purchasingpb.SupplierPayment{Id: v.ID, SupplierId: v.SupplierID, PaymentNumber: v.PaymentNumber,
		PaymentDate: v.PaymentDate, Currency: string(v.Currency), ExchangeRate: v.ExchangeRate, Amount: v.Amount,
		PaymentMethod: v.PaymentMethod, PaymentReference: v.PaymentReference, Status: string(v.Status),
		IdempotencyKey: v.IdempotencyKey, CreatedAt: v.CreatedAt, UpdatedAt: v.UpdatedAt, Allocations: items}
}

func toProtoOutstanding(v *domain.SupplierOutstandingBalance) *purchasingpb.SupplierOutstandingBalance {
	if v == nil {
		return nil
	}
	items := make([]*purchasingpb.SupplierInvoiceBalance, len(v.Invoices))
	for i, item := range v.Invoices {
		items[i] = &purchasingpb.SupplierInvoiceBalance{SupplierInvoiceId: item.SupplierInvoiceID, SupplierId: item.SupplierID,
			DocumentType: item.DocumentType, PointOfSale: item.PointOfSale, DocumentNumber: item.DocumentNumber,
			IssueDate: item.IssueDate, Currency: item.Currency, TotalAmount: item.TotalAmount,
			AllocatedAmount: item.AllocatedAmount, OutstandingAmount: item.OutstandingAmount}
	}
	return &purchasingpb.SupplierOutstandingBalance{SupplierId: v.SupplierID, Currency: v.Currency,
		TotalOutstanding: v.TotalOutstanding, Invoices: items}
}

func toProtoVATPosition(v *domain.MonthlyVATPosition) *purchasingpb.MonthlyVATPosition {
	if v == nil {
		return nil
	}
	lines := make([]*purchasingpb.VATPositionLine, len(v.Lines))
	for i, line := range v.Lines {
		lines[i] = &purchasingpb.VATPositionLine{TaxRate: line.TaxRate, SalesDebit: line.SalesDebit,
			PurchaseCredit: line.PurchaseCredit, NetVat: line.NetVAT}
	}
	return &purchasingpb.MonthlyVATPosition{Month: v.Month, SalesDebitTotal: v.SalesDebitTotal,
		PurchaseCreditTotal: v.PurchaseCreditTotal, NetVat: v.NetVAT, Lines: lines}
}

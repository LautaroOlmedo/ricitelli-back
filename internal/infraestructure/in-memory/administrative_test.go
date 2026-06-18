package repository

import (
	"context"
	"sync"
	"testing"

	customer_receipt "ricitelli-back/internal/domain/customer-receipt"
	product_inventory "ricitelli-back/internal/domain/product-inventory"
	purchasing "ricitelli-back/internal/domain/purchasing-administration"
	"ricitelli-back/internal/domain/remittance"
	sale_order "ricitelli-back/internal/domain/sale-order"
	sales_invoice "ricitelli-back/internal/domain/sales-invoice"
	sales_administration "ricitelli-back/internal/service/sales-administration"
	valueObject "ricitelli-back/internal/value-object"
)

func TestAdministrativeConcurrentIdempotency(t *testing.T) {
	ctx := context.Background()
	repo := NewInMemoryRepositoryMinimal()
	const workers = 32
	ids := make(chan string, workers)
	errs := make(chan error, workers)
	var wg sync.WaitGroup
	for range workers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			supplier, err := repo.CreateSupplier(ctx, purchasing.NewSupplierParams{
				TaxID: "30-999", SocialReason: "Concurrent Supplier", IdempotencyKey: "concurrent-supplier-key",
			})
			errs <- err
			ids <- supplier.ID
		}()
	}
	wg.Wait()
	close(ids)
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatal(err)
		}
	}
	var first string
	for id := range ids {
		if first == "" {
			first = id
		}
		if id != first {
			t.Fatalf("idempotent IDs differ: %q and %q", first, id)
		}
	}
}

func TestAdministrativeSalesIdempotencyAllocationsAndDispatch(t *testing.T) {
	ctx := context.Background()
	repo := NewInMemoryRepositoryMinimal()
	repo.SaleOrders = append(repo.SaleOrders, sale_order.ReconstitueSaleOrder(
		"order-1", "customer-1", sale_order.StatusConfirmed, nil,
		sale_order.CurrencyARS, sale_order.MarketDomestic, "", sale_order.SaleTypeRegular, "", true,
	))
	inventory, err := product_inventory.NewProductInventory("product-1", "SKU-1")
	if err != nil {
		t.Fatal(err)
	}
	if err := inventory.AddUndressed("production-1", 10, "system"); err != nil {
		t.Fatal(err)
	}
	if err := inventory.ConvertSVtoPT("dress-1", 10, "", "system"); err != nil {
		t.Fatal(err)
	}
	repo.ProductInventory = append(repo.ProductInventory, inventory)

	invoice, err := sales_invoice.NewSalesInvoice(sales_invoice.NewSalesInvoiceParams{
		CustomerID: "customer-1", SaleOrderID: "order-1", DocumentType: sales_invoice.DocumentTypeInvoice,
		PointOfSale: 1, DocumentNumber: 1, IssueDate: "2026-06-09", Currency: "ARS", ExchangeRate: "1",
		Subtotal: "100", TaxTotal: "21", TotalAmount: "121", IdempotencyKey: "invoice-key",
		Items: []sales_invoice.Item{{
			LineNumber: 1, ProductID: "product-1", Description: "Wine", Quantity: "1",
			UnitPrice: "100", TaxRate: "21", NetAmount: "100", TaxAmount: "21", TotalAmount: "121",
		}},
	})
	if err != nil {
		t.Fatal(err)
	}
	first, err := repo.SaveSalesInvoice(ctx, invoice)
	if err != nil {
		t.Fatal(err)
	}
	second, err := repo.SaveSalesInvoice(ctx, invoice)
	if err != nil || second.GetID() != first.GetID() {
		t.Fatalf("idempotent invoice = %q, %v; want %q", second.GetID(), err, first.GetID())
	}
	if _, err := repo.IssueSalesInvoice(ctx, first.GetID()); err != nil {
		t.Fatal(err)
	}

	receipt, err := customer_receipt.NewCustomerReceipt(customer_receipt.NewCustomerReceiptParams{
		CustomerID: "customer-1", ReceiptNumber: 1, ReceiptDate: "2026-06-09", Currency: "ARS",
		ExchangeRate: "1", Amount: "60", PaymentMethod: "TRANSFER", IdempotencyKey: "receipt-key",
		Allocations: []customer_receipt.Allocation{{SalesInvoiceID: first.GetID(), AllocatedAmount: "60"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	savedReceipt, err := repo.SaveCustomerReceipt(ctx, receipt)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := repo.PostCustomerReceipt(ctx, savedReceipt.GetID()); err != nil {
		t.Fatal(err)
	}
	balance, err := repo.GetSalesInvoiceOutstandingBalance(ctx, first.GetID())
	if err != nil || balance != "61" {
		t.Fatalf("balance = %q, %v; want 61", balance, err)
	}

	document, err := remittance.NewRemittance(remittance.NewRemittanceParams{
		CustomerID: "customer-1", SaleOrderID: "order-1", SalesInvoiceID: first.GetID(),
		PointOfSale: 1, DocumentNumber: 1, IssueDate: "2026-06-09", IdempotencyKey: "remittance-key",
		Items: []remittance.Item{{LineNumber: 1, ProductID: "product-1", Description: "Wine", Quantity: "2"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	savedDocument, err := repo.SaveRemittance(ctx, document)
	if err != nil {
		t.Fatal(err)
	}
	reloadedDocument, err := repo.GetRemittanceByID(ctx, savedDocument.GetID())
	if err != nil || reloadedDocument.GetSalesInvoiceID() != first.GetID() {
		t.Fatalf("persisted remittance invoice = %q, %v; want %q", reloadedDocument.GetSalesInvoiceID(), err, first.GetID())
	}
	for range 2 {
		if _, err := repo.ConfirmRemittance(ctx, savedDocument.GetID(), "user-1"); err != nil {
			t.Fatal(err)
		}
	}
	dispatches := 0
	for _, movement := range repo.ProductInventory[0].GetMovements() {
		if movement.MovementType == valueObject.ProductDispatched && movement.Reference == savedDocument.GetID() {
			dispatches++
		}
	}
	if dispatches != 1 {
		t.Fatalf("dispatches = %d, want 1", dispatches)
	}

	invoices, err := repo.ListSalesInvoices(ctx, sales_administration.SalesInvoiceFilter{
		CustomerID: "customer-1", SaleOrderID: "order-1", Status: string(sales_invoice.StatusPartiallyPaid),
	})
	if err != nil || len(invoices) != 1 {
		t.Fatalf("filtered invoices = %d, %v; want 1", len(invoices), err)
	}
	remittances, err := repo.ListRemittances(ctx, sales_administration.RemittanceFilter{
		CustomerID: "customer-1", SaleOrderID: "order-1", Status: string(remittance.StatusConfirmed),
	})
	if err != nil || len(remittances) != 1 {
		t.Fatalf("filtered remittances = %d, %v; want 1", len(remittances), err)
	}
	receipts, err := repo.ListCustomerReceipts(ctx, sales_administration.CustomerReceiptFilter{
		CustomerID: "customer-1", SaleOrderID: "order-1", Status: string(customer_receipt.StatusPosted),
	})
	if err != nil || len(receipts) != 1 {
		t.Fatalf("filtered receipts = %d, %v; want 1", len(receipts), err)
	}
}

func TestAdministrativePurchasingAllocationsAndVAT(t *testing.T) {
	ctx := context.Background()
	repo := NewInMemoryRepositoryMinimal()
	supplier, err := repo.CreateSupplier(ctx, purchasing.NewSupplierParams{
		TaxID: "30-123", SocialReason: "Supplier", IdempotencyKey: "supplier-key",
	})
	if err != nil {
		t.Fatal(err)
	}
	retry, err := repo.CreateSupplier(ctx, purchasing.NewSupplierParams{
		TaxID: "30-123", SocialReason: "Supplier", IdempotencyKey: "supplier-key",
	})
	if err != nil || retry.ID != supplier.ID {
		t.Fatalf("idempotent supplier = %q, %v; want %q", retry.ID, err, supplier.ID)
	}
	invoice, err := repo.CreateSupplierInvoice(ctx, purchasing.NewSupplierInvoiceParams{
		SupplierID: supplier.ID, DocumentType: purchasing.SupplierInvoiceTypeInvoice,
		PointOfSale: 1, DocumentNumber: 1, IssueDate: "2026-06-09", Currency: purchasing.CurrencyARS,
		ExchangeRate: "1", Subtotal: "100", TaxTotal: "21", TotalAmount: "121", IdempotencyKey: "supplier-invoice-key",
		Items: []purchasing.SupplierInvoiceItem{{
			LineNumber: 1, Description: "Bottles", Quantity: "1", Unit: "UNIT", UnitPrice: "100",
			TaxRate: "21", NetAmount: "100", TaxAmount: "21", TotalAmount: "121",
		}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := repo.UpdateSupplierInvoiceStatus(ctx, invoice.ID, purchasing.SupplierInvoiceIssued); err != nil {
		t.Fatal(err)
	}
	payment, err := repo.CreateSupplierPayment(ctx, purchasing.NewSupplierPaymentParams{
		SupplierID: supplier.ID, PaymentNumber: 1, PaymentDate: "2026-06-09", Currency: purchasing.CurrencyARS,
		ExchangeRate: "1", Amount: "60", PaymentMethod: "TRANSFER", IdempotencyKey: "payment-key",
		Allocations: []purchasing.SupplierPaymentAllocation{{SupplierInvoiceID: invoice.ID, AllocatedAmount: "60"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := repo.UpdateSupplierPaymentStatus(ctx, payment.ID, purchasing.SupplierPaymentPosted); err != nil {
		t.Fatal(err)
	}
	updatedInvoice, err := repo.GetSupplierInvoiceByID(ctx, invoice.ID)
	if err != nil || updatedInvoice.Status != purchasing.SupplierInvoicePartiallyPaid {
		t.Fatalf("invoice status = %q, %v; want PARTIALLY_PAID", updatedInvoice.Status, err)
	}
	balance, err := repo.GetSupplierOutstandingBalance(ctx, supplier.ID, purchasing.CurrencyARS)
	if err != nil || balance.TotalOutstanding != "61" {
		t.Fatalf("outstanding = %q, %v; want 61", balance.TotalOutstanding, err)
	}
	vat, err := repo.GetMonthlyVATPosition(ctx, "2026-06")
	if err != nil || vat.PurchaseCreditTotal != "21" {
		t.Fatalf("purchase VAT = %q, %v; want 21", vat.PurchaseCreditTotal, err)
	}
}

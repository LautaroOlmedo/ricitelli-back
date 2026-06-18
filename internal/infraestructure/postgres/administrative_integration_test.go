//go:build integration

package postgres

import (
	"context"
	"fmt"
	"math/big"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"

	customer_receipt "ricitelli-back/internal/domain/customer-receipt"
	purchasing "ricitelli-back/internal/domain/purchasing-administration"
	"ricitelli-back/internal/domain/remittance"
	sales_invoice "ricitelli-back/internal/domain/sales-invoice"
)

const composeDatabaseURL = "postgres://riccitelli:riccitelli@localhost:5432/riccitelli?sslmode=disable"

func TestAdministrativeIntegrationV1AAndV1B(t *testing.T) {
	ctx := context.Background()
	repo := openIntegrationRepository(t)
	defer repo.pool.Close()

	// Runtime migrations must be safe to execute repeatedly on an existing database.
	for i := 0; i < 2; i++ {
		if err := RunMigrations(repo.pool); err != nil {
			t.Fatalf("RunMigrations pass %d: %v", i+1, err)
		}
	}

	seed := seedAdministrativeIntegration(t, ctx, repo)
	initialVAT := vatTotals(t, ctx, repo, seed.month)

	invoice, err := sales_invoice.NewSalesInvoice(sales_invoice.NewSalesInvoiceParams{
		CustomerID: seed.customerID, SaleOrderID: seed.saleOrderID, DocumentType: sales_invoice.DocumentTypeInvoice,
		PointOfSale: seed.pointOfSale, DocumentNumber: seed.number, IssueDate: seed.date, Currency: "ARS", ExchangeRate: "1",
		Subtotal: "100", TaxTotal: "21", TotalAmount: "121", IdempotencyKey: seed.prefix + "-sales-invoice",
		Items: []sales_invoice.Item{{
			LineNumber: 1, ProductID: seed.productID, Description: "Integration wine", Quantity: "2",
			UnitPrice: "50", TaxRate: "21", NetAmount: "100", TaxAmount: "21", TotalAmount: "121",
		}},
	})
	if err != nil {
		t.Fatal(err)
	}
	movementsBefore := productMovementCount(t, ctx, repo, seed.inventoryID)
	savedInvoice, err := repo.SaveSalesInvoice(ctx, invoice)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := repo.IssueSalesInvoice(ctx, savedInvoice.GetID()); err != nil {
		t.Fatal(err)
	}
	if got := productMovementCount(t, ctx, repo, seed.inventoryID); got != movementsBefore {
		t.Fatalf("sales invoice changed stock movements: before=%d after=%d", movementsBefore, got)
	}

	document, err := remittance.NewRemittance(remittance.NewRemittanceParams{
		CustomerID: seed.customerID, SaleOrderID: seed.saleOrderID, SalesInvoiceID: savedInvoice.GetID(),
		PointOfSale: seed.pointOfSale, DocumentNumber: seed.number, IssueDate: seed.date,
		IdempotencyKey: seed.prefix + "-remittance",
		Items:          []remittance.Item{{LineNumber: 1, ProductID: seed.productID, Description: "Integration wine", Quantity: "2"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	savedDocument, err := repo.SaveRemittance(ctx, document)
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 2; i++ {
		if _, err := repo.ConfirmRemittance(ctx, savedDocument.GetID(), "integration-test"); err != nil {
			t.Fatalf("ConfirmRemittance pass %d: %v", i+1, err)
		}
	}
	var dispatches int
	if err := repo.pool.QueryRow(ctx, `SELECT COUNT(*) FROM product_inventory_movements
		WHERE product_inventory_id=$1 AND movement_type='PRODUCT_DISPATCHED' AND reference=$2`,
		seed.inventoryID, savedDocument.GetID()).Scan(&dispatches); err != nil {
		t.Fatal(err)
	}
	if dispatches != 1 {
		t.Fatalf("remittance dispatch movements = %d, want exactly 1", dispatches)
	}

	receipt, err := customer_receipt.NewCustomerReceipt(customer_receipt.NewCustomerReceiptParams{
		CustomerID: seed.customerID, ReceiptNumber: seed.number, ReceiptDate: seed.date, Currency: "ARS",
		ExchangeRate: "1", Amount: "60", PaymentMethod: "TRANSFER", IdempotencyKey: seed.prefix + "-receipt",
		Allocations: []customer_receipt.Allocation{{SalesInvoiceID: savedInvoice.GetID(), AllocatedAmount: "60"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	savedReceipt, err := repo.SaveCustomerReceipt(ctx, receipt)
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 2; i++ {
		if _, err := repo.PostCustomerReceipt(ctx, savedReceipt.GetID()); err != nil {
			t.Fatalf("PostCustomerReceipt pass %d: %v", i+1, err)
		}
	}
	if outstanding, err := repo.GetSalesInvoiceOutstandingBalance(ctx, savedInvoice.GetID()); err != nil || !decimalEqual(outstanding, "61") {
		t.Fatalf("sales invoice outstanding = %q, %v; want 61", outstanding, err)
	}

	supplier, err := repo.CreateSupplier(ctx, purchasing.NewSupplierParams{
		TaxID: seed.prefix[len("postgres-admin-it-"):], SocialReason: "Integration Supplier", IdempotencyKey: seed.prefix + "-supplier",
	})
	if err != nil {
		t.Fatal(err)
	}
	supplierInvoice, err := repo.CreateSupplierInvoice(ctx, purchasing.NewSupplierInvoiceParams{
		SupplierID: supplier.ID, DocumentType: purchasing.SupplierInvoiceTypeInvoice, PointOfSale: seed.pointOfSale,
		DocumentNumber: seed.number, IssueDate: seed.date, Currency: purchasing.CurrencyARS, ExchangeRate: "1",
		Subtotal: "100", TaxTotal: "21", TotalAmount: "121", IdempotencyKey: seed.prefix + "-supplier-invoice",
		Items: []purchasing.SupplierInvoiceItem{{
			LineNumber: 1, Description: "Integration bottles", Quantity: "1", Unit: "UNIT", UnitPrice: "100",
			TaxRate: "21", NetAmount: "100", TaxAmount: "21", TotalAmount: "121",
		}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := repo.UpdateSupplierInvoiceStatus(ctx, supplierInvoice.ID, purchasing.SupplierInvoiceIssued); err != nil {
		t.Fatal(err)
	}
	payment, err := repo.CreateSupplierPayment(ctx, purchasing.NewSupplierPaymentParams{
		SupplierID: supplier.ID, PaymentNumber: seed.number, PaymentDate: seed.date, Currency: purchasing.CurrencyARS,
		ExchangeRate: "1", Amount: "60", PaymentMethod: "TRANSFER", IdempotencyKey: seed.prefix + "-payment",
		Allocations: []purchasing.SupplierPaymentAllocation{{SupplierInvoiceID: supplierInvoice.ID, AllocatedAmount: "60"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 2; i++ {
		if _, err := repo.UpdateSupplierPaymentStatus(ctx, payment.ID, purchasing.SupplierPaymentPosted); err != nil {
			t.Fatalf("post supplier payment pass %d: %v", i+1, err)
		}
	}
	balance, err := repo.GetSupplierOutstandingBalance(ctx, supplier.ID, purchasing.CurrencyARS)
	if err != nil || !decimalEqual(balance.TotalOutstanding, "61") {
		t.Fatalf("supplier outstanding = %q, %v; want 61", balance.TotalOutstanding, err)
	}

	finalVAT := vatTotals(t, ctx, repo, seed.month)
	if !decimalEqual(decimalSubtract(finalVAT.sales, initialVAT.sales), "21") {
		t.Fatalf("sales VAT delta = %s, want 21", decimalSubtract(finalVAT.sales, initialVAT.sales))
	}
	if !decimalEqual(decimalSubtract(finalVAT.purchases, initialVAT.purchases), "21") {
		t.Fatalf("purchase VAT delta = %s, want 21", decimalSubtract(finalVAT.purchases, initialVAT.purchases))
	}
}

type integrationSeed struct {
	prefix, customerID, saleOrderID, productID, inventoryID, date, month string
	pointOfSale                                                          int32
	number                                                               int64
}

func openIntegrationRepository(t *testing.T) *Repository {
	t.Helper()
	databaseURL := os.Getenv("POSTGRES_INTEGRATION_URL")
	if databaseURL == "" {
		databaseURL = composeDatabaseURL
	}
	repo, err := NewRepository(databaseURL)
	if err != nil {
		t.Fatalf("open compose PostgreSQL (%s): %v", databaseURL, err)
	}
	return repo
}

func seedAdministrativeIntegration(t *testing.T, ctx context.Context, repo *Repository) integrationSeed {
	t.Helper()
	token := uuid.NewString()
	now := time.Now().UTC()
	number := now.UnixNano()
	seed := integrationSeed{
		prefix: "postgres-admin-it-" + token, customerID: uuid.NewString(), saleOrderID: uuid.NewString(),
		productID: uuid.NewString(), inventoryID: uuid.NewString(), date: now.Format(time.DateOnly), month: now.Format("2006-01"),
		pointOfSale: int32(number%1_000_000 + 1), number: number,
	}
	statements := []struct {
		sql  string
		args []any
	}{
		{`INSERT INTO customers(id,social_reason,market_type,customer_group,active,created_at) VALUES($1,$2,'EXTERNAL','RETAIL',true,$3)`,
			[]any{seed.customerID, seed.prefix, time.Now().UTC().Format(time.RFC3339Nano)}},
		{`INSERT INTO products(id,name,active) VALUES($1,$2,true)`, []any{seed.productID, seed.prefix}},
		{`INSERT INTO sale_orders(id,customer_id,status,currency,market,sale_type,active,created_at)
			VALUES($1,$2,'CONFIRMED','ARS','DOMESTIC','SALE',true,$3)`,
			[]any{seed.saleOrderID, seed.customerID, time.Now().UTC().Format(time.RFC3339Nano)}},
		{`INSERT INTO sale_order_items(sale_order_id,product_id,quantity,unit_price) VALUES($1,$2,2,50)`,
			[]any{seed.saleOrderID, seed.productID}},
		{`INSERT INTO product_inventories(id,product_id,sku) VALUES($1,$2,$3)`,
			[]any{seed.inventoryID, seed.productID, seed.prefix}},
	}
	for _, statement := range statements {
		if _, err := repo.pool.Exec(ctx, statement.sql, statement.args...); err != nil {
			t.Fatalf("seed integration row: %v", err)
		}
	}
	return seed
}

func productMovementCount(t *testing.T, ctx context.Context, repo *Repository, inventoryID string) int {
	t.Helper()
	var count int
	if err := repo.pool.QueryRow(ctx, `SELECT COUNT(*) FROM product_inventory_movements WHERE product_inventory_id=$1`, inventoryID).Scan(&count); err != nil {
		t.Fatal(err)
	}
	return count
}

type vatTotal struct{ sales, purchases string }

func vatTotals(t *testing.T, ctx context.Context, repo *Repository, month string) vatTotal {
	t.Helper()
	position, err := repo.GetMonthlyVATPosition(ctx, month)
	if err != nil {
		t.Fatal(err)
	}
	return vatTotal{sales: position.SalesDebitTotal, purchases: position.PurchaseCreditTotal}
}

func decimalEqual(left, right string) bool {
	return decimalRat(left).Cmp(decimalRat(right)) == 0
}

func decimalSubtract(left, right string) string {
	return new(big.Rat).Sub(decimalRat(left), decimalRat(right)).FloatString(4)
}

func decimalRat(value string) *big.Rat {
	result, ok := new(big.Rat).SetString(value)
	if !ok {
		panic(fmt.Sprintf("invalid decimal %q", value))
	}
	return result
}

package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	customer_receipt "ricitelli-back/internal/domain/customer-receipt"
	sales_invoice "ricitelli-back/internal/domain/sales-invoice"
	sales_administration "ricitelli-back/internal/service/sales-administration"
	valueObject "ricitelli-back/internal/value-object"

	"github.com/jackc/pgx/v5"
)

func (r *Repository) SaveCustomerReceipt(ctx context.Context, receipt customer_receipt.CustomerReceipt) (*customer_receipt.CustomerReceipt, error) {
	if existing, err := r.getCustomerReceiptByIdempotencyKey(ctx, receipt.GetIdempotencyKey()); err == nil {
		return existing, nil
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return nil, err
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)
	for _, allocation := range receipt.GetAllocations() {
		var customerID, status, documentType, currency string
		if err := tx.QueryRow(ctx, `SELECT customer_id::text, status, document_type, currency FROM sales_invoices WHERE id = $1`,
			allocation.SalesInvoiceID).Scan(&customerID, &status, &documentType, &currency); err != nil {
			return nil, fmt.Errorf("receipt allocation invoice: %w", err)
		}
		if customerID != receipt.GetCustomerID() {
			return nil, errors.New("receipt allocation invoice belongs to another customer")
		}
		if currency != receipt.GetCurrency() {
			return nil, errors.New("receipt allocation invoice uses another currency")
		}
		if status == "DRAFT" || status == "CANCELLED" || sales_invoice.DocumentType(documentType) == sales_invoice.DocumentTypeCredit {
			return nil, errors.New("receipt allocations require an issued invoice or debit")
		}
	}
	_, err = tx.Exec(ctx, `INSERT INTO customer_receipts
		(id, customer_id, receipt_number, receipt_date, currency, exchange_rate, amount, payment_method,
		 payment_reference, status, idempotency_key, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,NULLIF($9,''),'DRAFT',$10,$11,$12)`,
		receipt.GetID(), receipt.GetCustomerID(), receipt.GetReceiptNumber(), receipt.GetReceiptDate(), receipt.GetCurrency(),
		receipt.GetExchangeRate(), receipt.GetAmount(), receipt.GetPaymentMethod(), receipt.GetPaymentReference(),
		receipt.GetIdempotencyKey(), receipt.GetCreatedAt(), receipt.GetUpdatedAt())
	if err != nil {
		return nil, err
	}
	for _, allocation := range receipt.GetAllocations() {
		_, err = tx.Exec(ctx, `INSERT INTO customer_receipt_allocations
			(customer_receipt_id, sales_invoice_id, allocated_amount) VALUES ($1,$2,$3)`,
			receipt.GetID(), allocation.SalesInvoiceID, allocation.AllocatedAmount)
		if err != nil {
			return nil, err
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return r.GetCustomerReceiptByID(ctx, receipt.GetID())
}

func (r *Repository) GetCustomerReceiptByID(ctx context.Context, id string) (*customer_receipt.CustomerReceipt, error) {
	return r.loadCustomerReceipt(ctx, id, "")
}

func (r *Repository) ListCustomerReceipts(ctx context.Context, filter sales_administration.CustomerReceiptFilter) ([]customer_receipt.CustomerReceipt, error) {
	where, args := buildAdministrativeFilter(filter.CustomerID, "", filter.Status, "cr.customer_id", "", "cr.status")
	if filter.SaleOrderID != "" {
		args = append(args, filter.SaleOrderID)
		clause := fmt.Sprintf(`EXISTS (
			SELECT 1 FROM customer_receipt_allocations cra
			JOIN sales_invoices si ON si.id = cra.sales_invoice_id
			WHERE cra.customer_receipt_id = cr.id AND si.sale_order_id = $%d)`, len(args))
		if where == "" {
			where = " WHERE " + clause
		} else {
			where += " AND " + clause
		}
	}
	rows, err := r.pool.Query(ctx, `SELECT cr.id::text FROM customer_receipts cr`+where+` ORDER BY cr.created_at DESC`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	result := make([]customer_receipt.CustomerReceipt, 0, len(ids))
	for _, id := range ids {
		receipt, err := r.GetCustomerReceiptByID(ctx, id)
		if err != nil {
			return nil, err
		}
		result = append(result, *receipt)
	}
	return result, nil
}

func (r *Repository) getCustomerReceiptByIdempotencyKey(ctx context.Context, key string) (*customer_receipt.CustomerReceipt, error) {
	return r.loadCustomerReceipt(ctx, "", key)
}

// PostCustomerReceipt validates every allocation against the invoice's current
// posted outstanding balance under row locks before making the receipt immutable.
func (r *Repository) PostCustomerReceipt(ctx context.Context, id string) (*customer_receipt.CustomerReceipt, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	var status string
	if err := tx.QueryRow(ctx, `SELECT status FROM customer_receipts WHERE id = $1 FOR UPDATE`, id).Scan(&status); err != nil {
		return nil, err
	}
	if status == "POSTED" {
		if err := tx.Commit(ctx); err != nil {
			return nil, err
		}
		return r.GetCustomerReceiptByID(ctx, id)
	}
	if status != "DRAFT" {
		return nil, customer_receipt.ErrReceiptImmutable
	}
	rows, err := tx.Query(ctx, `SELECT sales_invoice_id::text, allocated_amount::text
		FROM customer_receipt_allocations WHERE customer_receipt_id = $1 ORDER BY sales_invoice_id FOR UPDATE`, id)
	if err != nil {
		return nil, err
	}
	var allocations []customer_receipt.Allocation
	for rows.Next() {
		var allocation customer_receipt.Allocation
		if err := rows.Scan(&allocation.SalesInvoiceID, &allocation.AllocatedAmount); err != nil {
			rows.Close()
			return nil, err
		}
		allocations = append(allocations, allocation)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return nil, err
	}
	rows.Close()
	for _, allocation := range allocations {
		outstanding, err := invoiceOutstandingInTx(ctx, tx, allocation.SalesInvoiceID)
		if err != nil {
			return nil, err
		}
		cmp, err := valueObject.CompareDecimal(allocation.AllocatedAmount, outstanding)
		if err != nil || cmp > 0 {
			return nil, errors.New("receipt allocation exceeds invoice outstanding balance")
		}
	}
	if _, err := tx.Exec(ctx, `UPDATE customer_receipts SET status = 'POSTED', updated_at = NOW() WHERE id = $1`, id); err != nil {
		return nil, err
	}
	for _, allocation := range allocations {
		outstanding, err := invoiceOutstandingInTx(ctx, tx, allocation.SalesInvoiceID)
		if err != nil {
			return nil, err
		}
		nextStatus := "PARTIALLY_PAID"
		if cmp, err := valueObject.CompareDecimal(outstanding, "0"); err == nil && cmp == 0 {
			nextStatus = "PAID"
		}
		if _, err := tx.Exec(ctx, `UPDATE sales_invoices SET status = $2, updated_at = NOW() WHERE id = $1`, allocation.SalesInvoiceID, nextStatus); err != nil {
			return nil, err
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return r.GetCustomerReceiptByID(ctx, id)
}

func invoiceOutstandingInTx(ctx context.Context, tx pgx.Tx, invoiceID string) (string, error) {
	var documentType, status, total, allocated string
	err := tx.QueryRow(ctx, `SELECT document_type, status, total_amount::text
		FROM sales_invoices WHERE id = $1 FOR UPDATE`, invoiceID).Scan(&documentType, &status, &total)
	if err != nil {
		return "", err
	}
	err = tx.QueryRow(ctx, `SELECT COALESCE(SUM(CASE WHEN cr.status = 'POSTED' THEN cra.allocated_amount ELSE 0 END), 0)::text
		FROM customer_receipt_allocations cra
		JOIN customer_receipts cr ON cr.id = cra.customer_receipt_id
		WHERE cra.sales_invoice_id = $1`, invoiceID).Scan(&allocated)
	if err != nil {
		return "", err
	}
	if status == "DRAFT" || status == "CANCELLED" || sales_invoice.DocumentType(documentType) == sales_invoice.DocumentTypeCredit {
		return "", errors.New("receipt allocation requires an issued invoice or debit")
	}
	return valueObject.SubtractDecimal(total, allocated)
}

func (r *Repository) loadCustomerReceipt(ctx context.Context, id, idempotencyKey string) (*customer_receipt.CustomerReceipt, error) {
	query := `SELECT id::text, customer_id::text, receipt_number, receipt_date::text, currency, exchange_rate::text,
		amount::text, payment_method, COALESCE(payment_reference,''), status, idempotency_key, created_at, updated_at
		FROM customer_receipts WHERE id = $1`
	arg := id
	if idempotencyKey != "" {
		query = `SELECT id::text, customer_id::text, receipt_number, receipt_date::text, currency, exchange_rate::text,
			amount::text, payment_method, COALESCE(payment_reference,''), status, idempotency_key, created_at, updated_at
			FROM customer_receipts WHERE idempotency_key = $1`
		arg = idempotencyKey
	}
	var receiptID, customerID, receiptDate, currency, exchangeRate, amount, paymentMethod, paymentReference string
	var storedStatus, key string
	var receiptNumber int64
	var createdAt, updatedAt time.Time
	if err := r.pool.QueryRow(ctx, query, arg).Scan(&receiptID, &customerID, &receiptNumber, &receiptDate, &currency,
		&exchangeRate, &amount, &paymentMethod, &paymentReference, &storedStatus, &key, &createdAt, &updatedAt); err != nil {
		return nil, err
	}
	allocations, err := r.loadCustomerReceiptAllocations(ctx, receiptID)
	if err != nil {
		return nil, err
	}
	status := customer_receipt.StatusDraft
	if storedStatus == "POSTED" {
		status = customer_receipt.StatusPosted
	}
	receipt := customer_receipt.ReconstituteCustomerReceipt(receiptID, customer_receipt.NewCustomerReceiptParams{
		CustomerID: customerID, ReceiptNumber: receiptNumber, ReceiptDate: receiptDate, Currency: currency,
		ExchangeRate: exchangeRate, Amount: amount, PaymentMethod: paymentMethod, PaymentReference: paymentReference,
		IdempotencyKey: key, Allocations: allocations,
	}, status, createdAt.UTC().Format(time.RFC3339Nano), updatedAt.UTC().Format(time.RFC3339Nano))
	return &receipt, nil
}

func (r *Repository) loadCustomerReceiptAllocations(ctx context.Context, receiptID string) ([]customer_receipt.Allocation, error) {
	rows, err := r.pool.Query(ctx, `SELECT sales_invoice_id::text, allocated_amount::text
		FROM customer_receipt_allocations WHERE customer_receipt_id = $1 ORDER BY sales_invoice_id`, receiptID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var allocations []customer_receipt.Allocation
	for rows.Next() {
		var allocation customer_receipt.Allocation
		if err := rows.Scan(&allocation.SalesInvoiceID, &allocation.AllocatedAmount); err != nil {
			return nil, err
		}
		allocations = append(allocations, allocation)
	}
	return allocations, rows.Err()
}

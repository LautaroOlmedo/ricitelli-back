package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	sales_invoice "ricitelli-back/internal/domain/sales-invoice"
	sales_administration "ricitelli-back/internal/service/sales-administration"
	valueObject "ricitelli-back/internal/value-object"

	"github.com/jackc/pgx/v5"
)

var _ sales_administration.Storage = (*Repository)(nil)

func (r *Repository) SaveSalesInvoice(ctx context.Context, invoice sales_invoice.SalesInvoice) (*sales_invoice.SalesInvoice, error) {
	if existing, err := r.getSalesInvoiceByIdempotencyKey(ctx, invoice.GetIdempotencyKey()); err == nil {
		return existing, nil
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return nil, err
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	var orderCustomerID string
	if err := tx.QueryRow(ctx, `SELECT customer_id::text FROM sale_orders WHERE id = $1`, invoice.GetSaleOrderID()).Scan(&orderCustomerID); err != nil {
		return nil, fmt.Errorf("sales invoice sale order: %w", err)
	}
	if orderCustomerID != invoice.GetCustomerID() {
		return nil, errors.New("sales invoice customer does not match sale order")
	}

	_, err = tx.Exec(ctx, `INSERT INTO sales_invoices
		(id, customer_id, sale_order_id, document_type, point_of_sale, document_number, issue_date, due_date,
		 currency, exchange_rate, subtotal, tax_total, total_amount, status, idempotency_key, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,NULLIF($8,'')::date,$9,$10,$11,$12,$13,$14,$15,$16,$17)`,
		invoice.GetID(), invoice.GetCustomerID(), invoice.GetSaleOrderID(), string(invoice.GetDocumentType()),
		invoice.GetPointOfSale(), invoice.GetDocumentNumber(), invoice.GetIssueDate(), invoice.GetDueDate(),
		invoice.GetCurrency(), invoice.GetExchangeRate(), invoice.GetSubtotal(), invoice.GetTaxTotal(),
		invoice.GetTotalAmount(), string(invoice.GetStatus()), invoice.GetIdempotencyKey(), invoice.GetCreatedAt(), invoice.GetUpdatedAt())
	if err != nil {
		return nil, err
	}
	for _, item := range invoice.GetItems() {
		_, err = tx.Exec(ctx, `INSERT INTO sales_invoice_items
			(sales_invoice_id, line_number, product_id, description, quantity, unit_price, tax_rate, net_amount, tax_amount, total_amount)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`,
			invoice.GetID(), item.LineNumber, item.ProductID, item.Description, item.Quantity, item.UnitPrice,
			item.TaxRate, item.NetAmount, item.TaxAmount, item.TotalAmount)
		if err != nil {
			return nil, err
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return r.GetSalesInvoiceByID(ctx, invoice.GetID())
}

func (r *Repository) GetSalesInvoiceByID(ctx context.Context, id string) (*sales_invoice.SalesInvoice, error) {
	return r.loadSalesInvoice(ctx, id, "")
}

func (r *Repository) ListSalesInvoices(ctx context.Context, filter sales_administration.SalesInvoiceFilter) ([]sales_invoice.SalesInvoice, error) {
	where, args := buildAdministrativeFilter(filter.CustomerID, filter.SaleOrderID, filter.Status, "customer_id", "sale_order_id", "status")
	rows, err := r.pool.Query(ctx, `SELECT id::text FROM sales_invoices`+where+` ORDER BY created_at DESC`, args...)
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
	result := make([]sales_invoice.SalesInvoice, 0, len(ids))
	for _, id := range ids {
		invoice, err := r.GetSalesInvoiceByID(ctx, id)
		if err != nil {
			return nil, err
		}
		result = append(result, *invoice)
	}
	return result, nil
}

func buildAdministrativeFilter(customerID, saleOrderID, status, customerColumn, orderColumn, statusColumn string) (string, []any) {
	clauses := make([]string, 0, 3)
	args := make([]any, 0, 3)
	add := func(column, value string) {
		if value != "" {
			args = append(args, value)
			clauses = append(clauses, fmt.Sprintf("%s = $%d", column, len(args)))
		}
	}
	add(customerColumn, customerID)
	add(orderColumn, saleOrderID)
	add(statusColumn, status)
	if len(clauses) == 0 {
		return "", args
	}
	return " WHERE " + strings.Join(clauses, " AND "), args
}

func (r *Repository) getSalesInvoiceByIdempotencyKey(ctx context.Context, key string) (*sales_invoice.SalesInvoice, error) {
	return r.loadSalesInvoice(ctx, "", key)
}

func (r *Repository) IssueSalesInvoice(ctx context.Context, id string) (*sales_invoice.SalesInvoice, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	var status string
	if err := tx.QueryRow(ctx, `SELECT status FROM sales_invoices WHERE id = $1 FOR UPDATE`, id).Scan(&status); err != nil {
		return nil, err
	}
	if status != string(sales_invoice.StatusDraft) && status != string(sales_invoice.StatusIssued) {
		return nil, sales_invoice.ErrInvoiceImmutable
	}
	if status == string(sales_invoice.StatusDraft) {
		if _, err := tx.Exec(ctx, `UPDATE sales_invoices SET status = 'ISSUED', updated_at = NOW() WHERE id = $1`, id); err != nil {
			return nil, err
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return r.GetSalesInvoiceByID(ctx, id)
}

func (r *Repository) GetSalesInvoiceOutstandingBalance(ctx context.Context, id string) (string, error) {
	var documentType, total, allocated string
	err := r.pool.QueryRow(ctx, `SELECT si.document_type, si.total_amount::text,
		COALESCE(SUM(CASE WHEN cr.status = 'POSTED' THEN cra.allocated_amount ELSE 0 END), 0)::text
		FROM sales_invoices si
		LEFT JOIN customer_receipt_allocations cra ON cra.sales_invoice_id = si.id
		LEFT JOIN customer_receipts cr ON cr.id = cra.customer_receipt_id
		WHERE si.id = $1
		GROUP BY si.document_type, si.total_amount`, id).Scan(&documentType, &total, &allocated)
	if err != nil {
		return "", err
	}
	if sales_invoice.DocumentType(documentType) == sales_invoice.DocumentTypeCredit {
		return valueObject.SubtractDecimal("0", total)
	}
	outstanding, err := valueObject.SubtractDecimal(total, allocated)
	if err != nil {
		return "", err
	}
	if cmp, _ := valueObject.CompareDecimal(outstanding, "0"); cmp < 0 {
		return "0", nil
	}
	return outstanding, nil
}

func (r *Repository) loadSalesInvoice(ctx context.Context, id, idempotencyKey string) (*sales_invoice.SalesInvoice, error) {
	query := `SELECT id::text, customer_id::text, sale_order_id::text, document_type, point_of_sale, document_number,
		issue_date::text, COALESCE(due_date::text,''), currency, exchange_rate::text, subtotal::text, tax_total::text,
		total_amount::text, status, idempotency_key, created_at, updated_at FROM sales_invoices WHERE id = $1`
	arg := id
	if idempotencyKey != "" {
		query = `SELECT id::text, customer_id::text, sale_order_id::text, document_type, point_of_sale, document_number,
			issue_date::text, COALESCE(due_date::text,''), currency, exchange_rate::text, subtotal::text, tax_total::text,
			total_amount::text, status, idempotency_key, created_at, updated_at FROM sales_invoices WHERE idempotency_key = $1`
		arg = idempotencyKey
	}
	var invoiceID, customerID, saleOrderID, documentType, issueDate, dueDate, currency, exchangeRate string
	var subtotal, taxTotal, totalAmount, storedStatus, key string
	var pointOfSale int32
	var documentNumber int64
	var createdAt, updatedAt time.Time
	if err := r.pool.QueryRow(ctx, query, arg).Scan(
		&invoiceID, &customerID, &saleOrderID, &documentType, &pointOfSale, &documentNumber, &issueDate, &dueDate,
		&currency, &exchangeRate, &subtotal, &taxTotal, &totalAmount, &storedStatus, &key, &createdAt, &updatedAt,
	); err != nil {
		return nil, err
	}
	items, err := r.loadSalesInvoiceItems(ctx, invoiceID)
	if err != nil {
		return nil, err
	}
	status := sales_invoice.Status(storedStatus)
	invoice := sales_invoice.ReconstituteSalesInvoice(invoiceID, sales_invoice.NewSalesInvoiceParams{
		CustomerID: customerID, SaleOrderID: saleOrderID, DocumentType: sales_invoice.DocumentType(documentType),
		PointOfSale: pointOfSale, DocumentNumber: documentNumber, IssueDate: issueDate, DueDate: dueDate,
		Currency: currency, ExchangeRate: exchangeRate, Subtotal: subtotal, TaxTotal: taxTotal, TotalAmount: totalAmount,
		IdempotencyKey: key, Items: items,
	}, status, createdAt.UTC().Format(time.RFC3339Nano), updatedAt.UTC().Format(time.RFC3339Nano))
	return &invoice, nil
}

func (r *Repository) loadSalesInvoiceItems(ctx context.Context, invoiceID string) ([]sales_invoice.Item, error) {
	rows, err := r.pool.Query(ctx, `SELECT line_number, COALESCE(product_id::text,''), description, quantity::text,
		unit_price::text, tax_rate::text, net_amount::text, tax_amount::text, total_amount::text
		FROM sales_invoice_items WHERE sales_invoice_id = $1 ORDER BY line_number`, invoiceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []sales_invoice.Item
	for rows.Next() {
		var item sales_invoice.Item
		if err := rows.Scan(&item.LineNumber, &item.ProductID, &item.Description, &item.Quantity, &item.UnitPrice,
			&item.TaxRate, &item.NetAmount, &item.TaxAmount, &item.TotalAmount); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

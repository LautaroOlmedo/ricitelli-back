package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"

	domain "ricitelli-back/internal/domain/purchasing-administration"
)

func (r *Repository) CreateSupplierInvoice(ctx context.Context, params domain.NewSupplierInvoiceParams) (domain.SupplierInvoice, error) {
	if existing, err := r.getSupplierInvoiceByIdempotencyKey(ctx, params.IdempotencyKey); err == nil {
		return *existing, nil
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return domain.SupplierInvoice{}, err
	}
	invoice, err := domain.NewSupplierInvoice(params)
	if err != nil {
		return domain.SupplierInvoice{}, err
	}
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return domain.SupplierInvoice{}, err
	}
	defer tx.Rollback(ctx)
	_, err = tx.Exec(ctx, `INSERT INTO supplier_invoices
		(id,supplier_id,supplier_quote_id,document_type,point_of_sale,document_number,issue_date,due_date,currency,
		 exchange_rate,subtotal,tax_total,total_amount,status,idempotency_key,created_at,updated_at)
		VALUES($1,$2,NULLIF($3,'')::uuid,$4,$5,$6,$7,NULLIF($8,'')::date,$9,$10::numeric,$11::numeric,$12::numeric,$13::numeric,$14,$15,$16,$17)`,
		invoice.ID, invoice.SupplierID, invoice.SupplierQuoteID, invoice.DocumentType, invoice.PointOfSale, invoice.DocumentNumber,
		invoice.IssueDate, invoice.DueDate, invoice.Currency, invoice.ExchangeRate, invoice.Subtotal, invoice.TaxTotal,
		invoice.TotalAmount, invoice.Status, invoice.IdempotencyKey, invoice.CreatedAt, invoice.UpdatedAt)
	if err != nil {
		return domain.SupplierInvoice{}, err
	}
	for _, item := range invoice.Items {
		_, err = tx.Exec(ctx, `INSERT INTO supplier_invoice_items
			(id,supplier_invoice_id,supplier_quote_item_id,line_number,dry_supply_id,description,quantity,unit,unit_price,tax_rate,net_amount,tax_amount,total_amount,created_at,updated_at)
			VALUES($1,$2,NULLIF($3,'')::uuid,$4,NULLIF($5,'')::uuid,$6,$7::numeric,$8,$9::numeric,$10::numeric,$11::numeric,$12::numeric,$13::numeric,$14,$15)`,
			item.ID, invoice.ID, item.SupplierQuoteItemID, item.LineNumber, item.DrySupplyID, item.Description, item.Quantity,
			item.Unit, item.UnitPrice, item.TaxRate, item.NetAmount, item.TaxAmount, item.TotalAmount, item.CreatedAt, item.UpdatedAt)
		if err != nil {
			return domain.SupplierInvoice{}, err
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return domain.SupplierInvoice{}, err
	}
	return invoice, nil
}

func (r *Repository) getSupplierInvoiceByIdempotencyKey(ctx context.Context, key string) (*domain.SupplierInvoice, error) {
	var id string
	if err := r.pool.QueryRow(ctx, `SELECT id::text FROM supplier_invoices WHERE idempotency_key=$1`, key).Scan(&id); err != nil {
		return nil, err
	}
	return r.loadSupplierInvoice(ctx, id)
}

func (r *Repository) GetSupplierInvoiceByID(ctx context.Context, id string) (*domain.SupplierInvoice, error) {
	return r.loadSupplierInvoice(ctx, id)
}

func (r *Repository) GetSupplierInvoices(ctx context.Context, supplierID string) ([]domain.SupplierInvoice, error) {
	query, args := `SELECT id::text FROM supplier_invoices`, []any{}
	if supplierID != "" {
		query, args = query+` WHERE supplier_id=$1`, []any{supplierID}
	}
	rows, err := r.pool.Query(ctx, query+` ORDER BY issue_date DESC,created_at DESC`, args...)
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
	result := make([]domain.SupplierInvoice, 0, len(ids))
	for _, id := range ids {
		value, err := r.loadSupplierInvoice(ctx, id)
		if err != nil {
			return nil, err
		}
		result = append(result, *value)
	}
	return result, rows.Err()
}

func (r *Repository) UpdateSupplierInvoiceStatus(ctx context.Context, id string, status domain.SupplierInvoiceStatus) (*domain.SupplierInvoice, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)
	var current domain.SupplierInvoiceStatus
	if err := tx.QueryRow(ctx, `SELECT status FROM supplier_invoices WHERE id=$1 FOR UPDATE`, id).Scan(&current); err != nil {
		return nil, err
	}
	if current == status {
		if err := tx.Commit(ctx); err != nil {
			return nil, err
		}
		return r.loadSupplierInvoice(ctx, id)
	}
	valid := (status == domain.SupplierInvoiceIssued && current == domain.SupplierInvoiceDraft) ||
		(status == domain.SupplierInvoiceCancelled && (current == domain.SupplierInvoiceDraft || current == domain.SupplierInvoiceIssued)) ||
		(status == domain.SupplierInvoicePartiallyPaid && current == domain.SupplierInvoiceIssued) ||
		(status == domain.SupplierInvoicePaid && (current == domain.SupplierInvoiceIssued || current == domain.SupplierInvoicePartiallyPaid))
	if !valid {
		return nil, errors.New("invalid supplier invoice status transition")
	}
	if _, err := tx.Exec(ctx, `UPDATE supplier_invoices SET status=$1,updated_at=NOW() WHERE id=$2`, status, id); err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return r.loadSupplierInvoice(ctx, id)
}

func (r *Repository) loadSupplierInvoice(ctx context.Context, id string) (*domain.SupplierInvoice, error) {
	var value domain.SupplierInvoice
	err := r.pool.QueryRow(ctx, `SELECT id::text,supplier_id::text,COALESCE(supplier_quote_id::text,''),document_type,
		point_of_sale,document_number,issue_date::text,COALESCE(due_date::text,''),currency,exchange_rate::text,
		subtotal::text,tax_total::text,total_amount::text,status,idempotency_key,created_at::text,updated_at::text
		FROM supplier_invoices WHERE id=$1`, id).
		Scan(&value.ID, &value.SupplierID, &value.SupplierQuoteID, &value.DocumentType, &value.PointOfSale, &value.DocumentNumber,
			&value.IssueDate, &value.DueDate, &value.Currency, &value.ExchangeRate, &value.Subtotal, &value.TaxTotal,
			&value.TotalAmount, &value.Status, &value.IdempotencyKey, &value.CreatedAt, &value.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, errors.New("supplier invoice not found")
	}
	if err != nil {
		return nil, err
	}
	rows, err := r.pool.Query(ctx, `SELECT id::text,supplier_invoice_id::text,COALESCE(supplier_quote_item_id::text,''),
		line_number,COALESCE(dry_supply_id::text,''),description,quantity::text,unit,unit_price::text,tax_rate::text,
		net_amount::text,tax_amount::text,total_amount::text,created_at::text,updated_at::text
		FROM supplier_invoice_items WHERE supplier_invoice_id=$1 ORDER BY line_number`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var item domain.SupplierInvoiceItem
		if err := rows.Scan(&item.ID, &item.SupplierInvoiceID, &item.SupplierQuoteItemID, &item.LineNumber, &item.DrySupplyID,
			&item.Description, &item.Quantity, &item.Unit, &item.UnitPrice, &item.TaxRate, &item.NetAmount, &item.TaxAmount,
			&item.TotalAmount, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, err
		}
		value.Items = append(value.Items, item)
	}
	return &value, rows.Err()
}

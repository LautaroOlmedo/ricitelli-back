package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"

	domain "ricitelli-back/internal/domain/purchasing-administration"
)

func (r *Repository) CreateSupplierQuote(ctx context.Context, params domain.NewSupplierQuoteParams) (domain.SupplierQuote, error) {
	if existing, err := r.getSupplierQuoteByIdempotencyKey(ctx, params.IdempotencyKey); err == nil {
		return *existing, nil
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return domain.SupplierQuote{}, err
	}
	quote, err := domain.NewSupplierQuote(params)
	if err != nil {
		return domain.SupplierQuote{}, err
	}
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return domain.SupplierQuote{}, err
	}
	defer tx.Rollback(ctx)
	_, err = tx.Exec(ctx, `INSERT INTO supplier_quotes
		(id,supplier_id,purchase_need_id,quote_number,quote_date,valid_until,currency,exchange_rate,subtotal,tax_total,total_amount,status,idempotency_key,created_at,updated_at)
		VALUES($1,$2,NULLIF($3,'')::uuid,$4,$5,NULLIF($6,'')::date,$7,$8::numeric,$9::numeric,$10::numeric,$11::numeric,$12,$13,$14,$15)`,
		quote.ID, quote.SupplierID, quote.PurchaseNeedID, quote.QuoteNumber, quote.QuoteDate, quote.ValidUntil,
		quote.Currency, quote.ExchangeRate, quote.Subtotal, quote.TaxTotal, quote.TotalAmount, quote.Status,
		quote.IdempotencyKey, quote.CreatedAt, quote.UpdatedAt)
	if err != nil {
		return domain.SupplierQuote{}, err
	}
	for _, item := range quote.Items {
		_, err = tx.Exec(ctx, `INSERT INTO supplier_quote_items
			(id,supplier_quote_id,purchase_need_item_id,line_number,dry_supply_id,description,quantity,unit,unit_price,tax_rate,net_amount,tax_amount,total_amount,created_at,updated_at)
			VALUES($1,$2,NULLIF($3,'')::uuid,$4,NULLIF($5,'')::uuid,$6,$7::numeric,$8,$9::numeric,$10::numeric,$11::numeric,$12::numeric,$13::numeric,$14,$15)`,
			item.ID, quote.ID, item.PurchaseNeedItemID, item.LineNumber, item.DrySupplyID, item.Description, item.Quantity,
			item.Unit, item.UnitPrice, item.TaxRate, item.NetAmount, item.TaxAmount, item.TotalAmount, item.CreatedAt, item.UpdatedAt)
		if err != nil {
			return domain.SupplierQuote{}, err
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return domain.SupplierQuote{}, err
	}
	return quote, nil
}

func (r *Repository) getSupplierQuoteByIdempotencyKey(ctx context.Context, key string) (*domain.SupplierQuote, error) {
	var id string
	if err := r.pool.QueryRow(ctx, `SELECT id::text FROM supplier_quotes WHERE idempotency_key=$1`, key).Scan(&id); err != nil {
		return nil, err
	}
	return r.loadSupplierQuote(ctx, id)
}

func (r *Repository) GetSupplierQuoteByID(ctx context.Context, id string) (*domain.SupplierQuote, error) {
	return r.loadSupplierQuote(ctx, id)
}

func (r *Repository) GetSupplierQuotes(ctx context.Context, supplierID string) ([]domain.SupplierQuote, error) {
	query, args := `SELECT id::text FROM supplier_quotes`, []any{}
	if supplierID != "" {
		query, args = query+` WHERE supplier_id=$1`, []any{supplierID}
	}
	rows, err := r.pool.Query(ctx, query+` ORDER BY quote_date DESC,created_at DESC`, args...)
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
	result := make([]domain.SupplierQuote, 0, len(ids))
	for _, id := range ids {
		value, err := r.loadSupplierQuote(ctx, id)
		if err != nil {
			return nil, err
		}
		result = append(result, *value)
	}
	return result, rows.Err()
}

func (r *Repository) UpdateSupplierQuoteStatus(ctx context.Context, id string, status domain.SupplierQuoteStatus) (*domain.SupplierQuote, error) {
	tag, err := r.pool.Exec(ctx, `UPDATE supplier_quotes SET status=$1,updated_at=NOW() WHERE id=$2`, status, id)
	if err != nil {
		return nil, err
	}
	if tag.RowsAffected() == 0 {
		return nil, errors.New("supplier quote not found")
	}
	return r.loadSupplierQuote(ctx, id)
}

func (r *Repository) loadSupplierQuote(ctx context.Context, id string) (*domain.SupplierQuote, error) {
	var value domain.SupplierQuote
	err := r.pool.QueryRow(ctx, `SELECT id::text,supplier_id::text,COALESCE(purchase_need_id::text,''),quote_number,
		quote_date::text,COALESCE(valid_until::text,''),currency,exchange_rate::text,subtotal::text,tax_total::text,
		total_amount::text,status,idempotency_key,created_at::text,updated_at::text FROM supplier_quotes WHERE id=$1`, id).
		Scan(&value.ID, &value.SupplierID, &value.PurchaseNeedID, &value.QuoteNumber, &value.QuoteDate, &value.ValidUntil,
			&value.Currency, &value.ExchangeRate, &value.Subtotal, &value.TaxTotal, &value.TotalAmount, &value.Status,
			&value.IdempotencyKey, &value.CreatedAt, &value.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, errors.New("supplier quote not found")
	}
	if err != nil {
		return nil, err
	}
	rows, err := r.pool.Query(ctx, `SELECT id::text,supplier_quote_id::text,COALESCE(purchase_need_item_id::text,''),
		line_number,COALESCE(dry_supply_id::text,''),description,quantity::text,unit,unit_price::text,tax_rate::text,
		net_amount::text,tax_amount::text,total_amount::text,created_at::text,updated_at::text
		FROM supplier_quote_items WHERE supplier_quote_id=$1 ORDER BY line_number`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var item domain.SupplierQuoteItem
		if err := rows.Scan(&item.ID, &item.SupplierQuoteID, &item.PurchaseNeedItemID, &item.LineNumber, &item.DrySupplyID,
			&item.Description, &item.Quantity, &item.Unit, &item.UnitPrice, &item.TaxRate, &item.NetAmount, &item.TaxAmount,
			&item.TotalAmount, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, err
		}
		value.Items = append(value.Items, item)
	}
	return &value, rows.Err()
}

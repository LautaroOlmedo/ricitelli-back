package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"

	domain "ricitelli-back/internal/domain/purchasing-administration"
)

func (r *Repository) CreateSupplier(ctx context.Context, params domain.NewSupplierParams) (domain.Supplier, error) {
	if existing, err := r.getSupplierByIdempotencyKey(ctx, params.IdempotencyKey); err == nil {
		return *existing, nil
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return domain.Supplier{}, err
	}
	supplier, err := domain.NewSupplier(params)
	if err != nil {
		return domain.Supplier{}, err
	}
	_, err = r.pool.Exec(ctx, `INSERT INTO suppliers
		(id,tax_id,social_reason,trade_name,email,phone,address,active,idempotency_key,created_at,updated_at)
		VALUES ($1,$2,$3,NULLIF($4,''),NULLIF($5,''),NULLIF($6,''),NULLIF($7,''),$8,$9,$10,$11)`,
		supplier.ID, supplier.TaxID, supplier.SocialReason, supplier.TradeName, supplier.Email, supplier.Phone,
		supplier.Address, supplier.Active, supplier.IdempotencyKey, supplier.CreatedAt, supplier.UpdatedAt)
	if err != nil {
		return domain.Supplier{}, err
	}
	return supplier, nil
}

func (r *Repository) getSupplierByIdempotencyKey(ctx context.Context, key string) (*domain.Supplier, error) {
	return scanSupplier(r.pool.QueryRow(ctx, `SELECT id::text,tax_id,social_reason,COALESCE(trade_name,''),
		COALESCE(email,''),COALESCE(phone,''),COALESCE(address,''),active,idempotency_key,created_at::text,updated_at::text
		FROM suppliers WHERE idempotency_key=$1`, key))
}

func (r *Repository) GetSupplierByID(ctx context.Context, id string) (*domain.Supplier, error) {
	return scanSupplier(r.pool.QueryRow(ctx, `SELECT id::text,tax_id,social_reason,COALESCE(trade_name,''),
		COALESCE(email,''),COALESCE(phone,''),COALESCE(address,''),active,idempotency_key,created_at::text,updated_at::text
		FROM suppliers WHERE id=$1`, id))
}

func (r *Repository) GetSuppliers(ctx context.Context, includeInactive bool) ([]domain.Supplier, error) {
	query := `SELECT id::text,tax_id,social_reason,COALESCE(trade_name,''),COALESCE(email,''),COALESCE(phone,''),
		COALESCE(address,''),active,idempotency_key,created_at::text,updated_at::text FROM suppliers`
	if !includeInactive {
		query += ` WHERE active=true`
	}
	query += ` ORDER BY social_reason`
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []domain.Supplier
	for rows.Next() {
		supplier, err := scanSupplier(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, *supplier)
	}
	return result, rows.Err()
}

func (r *Repository) DeactivateSupplier(ctx context.Context, id string) (*domain.Supplier, error) {
	return scanSupplier(r.pool.QueryRow(ctx, `UPDATE suppliers SET active=false,updated_at=NOW() WHERE id=$1 RETURNING
		id::text,tax_id,social_reason,COALESCE(trade_name,''),COALESCE(email,''),COALESCE(phone,''),
		COALESCE(address,''),active,idempotency_key,created_at::text,updated_at::text`, id))
}

func (r *Repository) UpdateSupplier(ctx context.Context, id string, params domain.UpdateSupplierParams) (*domain.Supplier, error) {
	supplier, err := r.GetSupplierByID(ctx, id)
	if err != nil {
		return nil, err
	}
	supplier.Update(params)
	return scanSupplier(r.pool.QueryRow(ctx, `UPDATE suppliers SET tax_id=$1,social_reason=$2,
		trade_name=NULLIF($3,''),email=NULLIF($4,''),phone=NULLIF($5,''),address=NULLIF($6,''),updated_at=$7
		WHERE id=$8 RETURNING id::text,tax_id,social_reason,COALESCE(trade_name,''),COALESCE(email,''),
		COALESCE(phone,''),COALESCE(address,''),active,idempotency_key,created_at::text,updated_at::text`,
		supplier.TaxID, supplier.SocialReason, supplier.TradeName, supplier.Email, supplier.Phone, supplier.Address,
		supplier.UpdatedAt, id))
}

type supplierScanner interface{ Scan(...any) error }

func scanSupplier(row supplierScanner) (*domain.Supplier, error) {
	var value domain.Supplier
	err := row.Scan(&value.ID, &value.TaxID, &value.SocialReason, &value.TradeName, &value.Email, &value.Phone,
		&value.Address, &value.Active, &value.IdempotencyKey, &value.CreatedAt, &value.UpdatedAt)
	return &value, err
}

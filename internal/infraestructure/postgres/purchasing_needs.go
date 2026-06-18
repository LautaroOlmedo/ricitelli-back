package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"

	domain "ricitelli-back/internal/domain/purchasing-administration"
)

func (r *Repository) CreatePurchaseNeed(ctx context.Context, params domain.NewPurchaseNeedParams) (domain.PurchaseNeed, error) {
	if existing, err := r.getPurchaseNeedByIdempotencyKey(ctx, params.IdempotencyKey); err == nil {
		return *existing, nil
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return domain.PurchaseNeed{}, err
	}
	need, err := domain.NewPurchaseNeed(params)
	if err != nil {
		return domain.PurchaseNeed{}, err
	}
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return domain.PurchaseNeed{}, err
	}
	defer tx.Rollback(ctx)
	_, err = tx.Exec(ctx, `INSERT INTO purchase_needs
		(id,need_number,requested_date,required_by_date,status,notes,idempotency_key,created_at,updated_at)
		VALUES($1,$2,$3,NULLIF($4,'')::date,$5,NULLIF($6,''),$7,$8,$9)`,
		need.ID, need.NeedNumber, need.RequestedDate, need.RequiredByDate, need.Status, need.Notes,
		need.IdempotencyKey, need.CreatedAt, need.UpdatedAt)
	if err != nil {
		return domain.PurchaseNeed{}, err
	}
	for _, item := range need.Items {
		_, err = tx.Exec(ctx, `INSERT INTO purchase_need_items
			(id,purchase_need_id,line_number,dry_supply_id,description,quantity,unit,created_at,updated_at)
			VALUES($1,$2,$3,NULLIF($4,'')::uuid,$5,$6::numeric,$7,$8,$9)`,
			item.ID, need.ID, item.LineNumber, item.DrySupplyID, item.Description, item.Quantity, item.Unit, item.CreatedAt, item.UpdatedAt)
		if err != nil {
			return domain.PurchaseNeed{}, err
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return domain.PurchaseNeed{}, err
	}
	return need, nil
}

func (r *Repository) getPurchaseNeedByIdempotencyKey(ctx context.Context, key string) (*domain.PurchaseNeed, error) {
	var id string
	if err := r.pool.QueryRow(ctx, `SELECT id::text FROM purchase_needs WHERE idempotency_key=$1`, key).Scan(&id); err != nil {
		return nil, err
	}
	return r.loadPurchaseNeed(ctx, id)
}

func (r *Repository) GetPurchaseNeedByID(ctx context.Context, id string) (*domain.PurchaseNeed, error) {
	return r.loadPurchaseNeed(ctx, id)
}

func (r *Repository) GetPurchaseNeeds(ctx context.Context) ([]domain.PurchaseNeed, error) {
	rows, err := r.pool.Query(ctx, `SELECT id::text FROM purchase_needs ORDER BY requested_date DESC,need_number DESC`)
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
	result := make([]domain.PurchaseNeed, 0, len(ids))
	for _, id := range ids {
		value, err := r.loadPurchaseNeed(ctx, id)
		if err != nil {
			return nil, err
		}
		result = append(result, *value)
	}
	return result, rows.Err()
}

func (r *Repository) UpdatePurchaseNeedStatus(ctx context.Context, id string, status domain.PurchaseNeedStatus) (*domain.PurchaseNeed, error) {
	tag, err := r.pool.Exec(ctx, `UPDATE purchase_needs SET status=$1,updated_at=NOW() WHERE id=$2`, status, id)
	if err != nil {
		return nil, err
	}
	if tag.RowsAffected() == 0 {
		return nil, errors.New("purchase need not found")
	}
	return r.loadPurchaseNeed(ctx, id)
}

func (r *Repository) loadPurchaseNeed(ctx context.Context, id string) (*domain.PurchaseNeed, error) {
	var value domain.PurchaseNeed
	err := r.pool.QueryRow(ctx, `SELECT id::text,need_number,requested_date::text,COALESCE(required_by_date::text,''),
		status,COALESCE(notes,''),idempotency_key,created_at::text,updated_at::text FROM purchase_needs WHERE id=$1`, id).
		Scan(&value.ID, &value.NeedNumber, &value.RequestedDate, &value.RequiredByDate, &value.Status, &value.Notes,
			&value.IdempotencyKey, &value.CreatedAt, &value.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, errors.New("purchase need not found")
	}
	if err != nil {
		return nil, err
	}
	rows, err := r.pool.Query(ctx, `SELECT id::text,purchase_need_id::text,line_number,COALESCE(dry_supply_id::text,''),
		description,quantity::text,unit,created_at::text,updated_at::text FROM purchase_need_items
		WHERE purchase_need_id=$1 ORDER BY line_number`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var item domain.PurchaseNeedItem
		if err := rows.Scan(&item.ID, &item.PurchaseNeedID, &item.LineNumber, &item.DrySupplyID, &item.Description,
			&item.Quantity, &item.Unit, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, err
		}
		value.Items = append(value.Items, item)
	}
	return &value, rows.Err()
}

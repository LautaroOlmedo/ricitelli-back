package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"

	domain "ricitelli-back/internal/domain/purchasing-administration"
)

func (r *Repository) CreateSupplierPayment(ctx context.Context, params domain.NewSupplierPaymentParams) (domain.SupplierPayment, error) {
	if existing, err := r.getSupplierPaymentByIdempotencyKey(ctx, params.IdempotencyKey); err == nil {
		return *existing, nil
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return domain.SupplierPayment{}, err
	}
	payment, err := domain.NewSupplierPayment(params)
	if err != nil {
		return domain.SupplierPayment{}, err
	}
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return domain.SupplierPayment{}, err
	}
	defer tx.Rollback(ctx)
	_, err = tx.Exec(ctx, `INSERT INTO supplier_payments
		(id,supplier_id,payment_number,payment_date,currency,exchange_rate,amount,payment_method,payment_reference,status,idempotency_key,created_at,updated_at)
		VALUES($1,$2,$3,$4,$5,$6::numeric,$7::numeric,$8,NULLIF($9,''),$10,$11,$12,$13)`,
		payment.ID, payment.SupplierID, payment.PaymentNumber, payment.PaymentDate, payment.Currency, payment.ExchangeRate,
		payment.Amount, payment.PaymentMethod, payment.PaymentReference, payment.Status, payment.IdempotencyKey,
		payment.CreatedAt, payment.UpdatedAt)
	if err != nil {
		return domain.SupplierPayment{}, err
	}
	for _, allocation := range payment.Allocations {
		_, err = tx.Exec(ctx, `INSERT INTO supplier_payment_allocations
			(id,supplier_payment_id,supplier_invoice_id,allocated_amount,created_at,updated_at)
			VALUES($1,$2,$3,$4::numeric,$5,$6)`,
			allocation.ID, payment.ID, allocation.SupplierInvoiceID, allocation.AllocatedAmount,
			allocation.CreatedAt, allocation.UpdatedAt)
		if err != nil {
			return domain.SupplierPayment{}, err
		}
	}
	if err := validateSupplierPaymentAllocations(ctx, tx, payment.ID); err != nil {
		return domain.SupplierPayment{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return domain.SupplierPayment{}, err
	}
	return payment, nil
}

func (r *Repository) getSupplierPaymentByIdempotencyKey(ctx context.Context, key string) (*domain.SupplierPayment, error) {
	var id string
	if err := r.pool.QueryRow(ctx, `SELECT id::text FROM supplier_payments WHERE idempotency_key=$1`, key).Scan(&id); err != nil {
		return nil, err
	}
	return r.loadSupplierPayment(ctx, id)
}

func (r *Repository) GetSupplierPaymentByID(ctx context.Context, id string) (*domain.SupplierPayment, error) {
	return r.loadSupplierPayment(ctx, id)
}

func (r *Repository) GetSupplierPayments(ctx context.Context, supplierID string) ([]domain.SupplierPayment, error) {
	query, args := `SELECT id::text FROM supplier_payments`, []any{}
	if supplierID != "" {
		query, args = query+` WHERE supplier_id=$1`, []any{supplierID}
	}
	rows, err := r.pool.Query(ctx, query+` ORDER BY payment_date DESC,created_at DESC`, args...)
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
	result := make([]domain.SupplierPayment, 0, len(ids))
	for _, id := range ids {
		value, err := r.loadSupplierPayment(ctx, id)
		if err != nil {
			return nil, err
		}
		result = append(result, *value)
	}
	return result, rows.Err()
}

func (r *Repository) UpdateSupplierPaymentStatus(ctx context.Context, id string, status domain.SupplierPaymentStatus) (*domain.SupplierPayment, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)
	var current domain.SupplierPaymentStatus
	if err := tx.QueryRow(ctx, `SELECT status FROM supplier_payments WHERE id=$1 FOR UPDATE`, id).Scan(&current); err != nil {
		return nil, err
	}
	if status != domain.SupplierPaymentPosted && status != domain.SupplierPaymentVoided {
		return nil, errors.New("invalid supplier payment status transition")
	}
	if current == status {
		if err := tx.Commit(ctx); err != nil {
			return nil, err
		}
		return r.loadSupplierPayment(ctx, id)
	}
	if (status == domain.SupplierPaymentPosted && current != domain.SupplierPaymentDraft) ||
		(status == domain.SupplierPaymentVoided && current != domain.SupplierPaymentPosted) {
		return nil, errors.New("invalid supplier payment status transition")
	}
	if status == domain.SupplierPaymentPosted {
		if err := lockSupplierPaymentInvoices(ctx, tx, id); err != nil {
			return nil, err
		}
		if err := validateSupplierPaymentAllocations(ctx, tx, id); err != nil {
			return nil, err
		}
	}
	tag, err := tx.Exec(ctx, `UPDATE supplier_payments SET status=$1,updated_at=NOW() WHERE id=$2`, status, id)
	if err != nil {
		return nil, err
	}
	if tag.RowsAffected() == 0 {
		return nil, errors.New("supplier payment not found")
	}
	_, err = tx.Exec(ctx, `UPDATE supplier_invoices si SET status = CASE
			WHEN si.status='CANCELLED' THEN si.status
			WHEN balances.allocated_amount >= si.total_amount THEN 'PAID'
			WHEN balances.allocated_amount > 0 THEN 'PARTIALLY_PAID'
			ELSE 'ISSUED' END,
			updated_at=NOW()
		FROM (
			SELECT spa.supplier_invoice_id,
				COALESCE(SUM(CASE WHEN sp.status='POSTED' THEN spa.allocated_amount ELSE 0 END),0) allocated_amount
			FROM supplier_payment_allocations spa
			JOIN supplier_payments sp ON sp.id=spa.supplier_payment_id
			WHERE spa.supplier_invoice_id IN (
				SELECT supplier_invoice_id FROM supplier_payment_allocations WHERE supplier_payment_id=$1)
			GROUP BY spa.supplier_invoice_id
		) balances
		WHERE si.id=balances.supplier_invoice_id AND si.status <> 'DRAFT'`, id)
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return r.loadSupplierPayment(ctx, id)
}

func lockSupplierPaymentInvoices(ctx context.Context, tx pgx.Tx, paymentID string) error {
	rows, err := tx.Query(ctx, `SELECT invoice.id
		FROM supplier_invoices invoice
		JOIN supplier_payment_allocations allocation ON allocation.supplier_invoice_id=invoice.id
		WHERE allocation.supplier_payment_id=$1
		ORDER BY invoice.id FOR UPDATE`, paymentID)
	if err != nil {
		return err
	}
	for rows.Next() {
		var ignored string
		if err := rows.Scan(&ignored); err != nil {
			rows.Close()
			return err
		}
	}
	rows.Close()
	return rows.Err()
}

func validateSupplierPaymentAllocations(ctx context.Context, tx pgx.Tx, paymentID string) error {
	var invalid int
	err := tx.QueryRow(ctx, `SELECT COUNT(*)
		FROM supplier_payment_allocations current_allocation
		JOIN supplier_payments current_payment ON current_payment.id=current_allocation.supplier_payment_id
		JOIN supplier_invoices invoice ON invoice.id=current_allocation.supplier_invoice_id
		LEFT JOIN LATERAL (
			SELECT COALESCE(SUM(other_allocation.allocated_amount),0) amount
			FROM supplier_payment_allocations other_allocation
			JOIN supplier_payments other_payment ON other_payment.id=other_allocation.supplier_payment_id
			WHERE other_allocation.supplier_invoice_id=invoice.id
			  AND other_payment.status='POSTED'
			  AND other_payment.id<>current_payment.id
		) posted ON true
		WHERE current_allocation.supplier_payment_id=$1
		  AND (invoice.supplier_id<>current_payment.supplier_id
		    OR invoice.currency<>current_payment.currency
		    OR invoice.document_type='CREDIT_NOTE'
		    OR invoice.status NOT IN ('ISSUED','PARTIALLY_PAID')
		    OR current_allocation.allocated_amount > invoice.total_amount-posted.amount)`, paymentID).Scan(&invalid)
	if err != nil {
		return err
	}
	if invalid > 0 {
		return errors.New("supplier payment contains an invalid or excessive invoice allocation")
	}
	return nil
}

func (r *Repository) GetSupplierOutstandingBalance(ctx context.Context, supplierID string, currency domain.Currency) (*domain.SupplierOutstandingBalance, error) {
	rows, err := r.pool.Query(ctx, `SELECT si.id::text,si.supplier_id::text,si.document_type,si.point_of_sale,si.document_number,
		si.issue_date::text,si.currency,
		(CASE WHEN si.document_type='CREDIT_NOTE' THEN -si.total_amount ELSE si.total_amount END)::text total_amount,
		COALESCE(SUM(CASE WHEN sp.status='POSTED' THEN spa.allocated_amount ELSE 0 END),0)::text allocated_amount,
		((CASE WHEN si.document_type='CREDIT_NOTE' THEN -si.total_amount ELSE si.total_amount END) -
		 COALESCE(SUM(CASE WHEN sp.status='POSTED' THEN spa.allocated_amount ELSE 0 END),0))::text outstanding_amount
		FROM supplier_invoices si
		LEFT JOIN supplier_payment_allocations spa ON spa.supplier_invoice_id=si.id
		LEFT JOIN supplier_payments sp ON sp.id=spa.supplier_payment_id
		WHERE si.supplier_id=$1 AND si.currency=$2 AND si.status IN ('ISSUED','PARTIALLY_PAID','PAID')
		GROUP BY si.id ORDER BY si.issue_date,si.document_number`, supplierID, currency)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := domain.SupplierOutstandingBalance{SupplierID: supplierID, Currency: string(currency), TotalOutstanding: "0"}
	for rows.Next() {
		var line domain.SupplierInvoiceBalance
		if err := rows.Scan(&line.SupplierInvoiceID, &line.SupplierID, &line.DocumentType, &line.PointOfSale,
			&line.DocumentNumber, &line.IssueDate, &line.Currency, &line.TotalAmount, &line.AllocatedAmount,
			&line.OutstandingAmount); err != nil {
			return nil, err
		}
		result.Invoices = append(result.Invoices, line)
		result.TotalOutstanding, err = domain.AddDecimals(result.TotalOutstanding, line.OutstandingAmount)
		if err != nil {
			return nil, err
		}
	}
	return &result, rows.Err()
}

func (r *Repository) loadSupplierPayment(ctx context.Context, id string) (*domain.SupplierPayment, error) {
	var value domain.SupplierPayment
	err := r.pool.QueryRow(ctx, `SELECT id::text,supplier_id::text,payment_number,payment_date::text,currency,
		exchange_rate::text,amount::text,payment_method,COALESCE(payment_reference,''),status,idempotency_key,
		created_at::text,updated_at::text FROM supplier_payments WHERE id=$1`, id).
		Scan(&value.ID, &value.SupplierID, &value.PaymentNumber, &value.PaymentDate, &value.Currency, &value.ExchangeRate,
			&value.Amount, &value.PaymentMethod, &value.PaymentReference, &value.Status, &value.IdempotencyKey,
			&value.CreatedAt, &value.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, errors.New("supplier payment not found")
	}
	if err != nil {
		return nil, err
	}
	rows, err := r.pool.Query(ctx, `SELECT id::text,supplier_payment_id::text,supplier_invoice_id::text,
		allocated_amount::text,created_at::text,updated_at::text FROM supplier_payment_allocations
		WHERE supplier_payment_id=$1 ORDER BY created_at,id`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var item domain.SupplierPaymentAllocation
		if err := rows.Scan(&item.ID, &item.SupplierPaymentID, &item.SupplierInvoiceID, &item.AllocatedAmount,
			&item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, err
		}
		value.Allocations = append(value.Allocations, item)
	}
	return &value, rows.Err()
}

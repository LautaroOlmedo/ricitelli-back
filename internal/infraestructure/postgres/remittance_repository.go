package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"ricitelli-back/internal/domain/remittance"
	sales_administration "ricitelli-back/internal/service/sales-administration"
	valueObject "ricitelli-back/internal/value-object"

	"github.com/jackc/pgx/v5"
)

func (r *Repository) SaveRemittance(ctx context.Context, document remittance.Remittance) (*remittance.Remittance, error) {
	if existing, err := r.getRemittanceByIdempotencyKey(ctx, document.GetIdempotencyKey()); err == nil {
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
	if err := tx.QueryRow(ctx, `SELECT customer_id::text FROM sale_orders WHERE id = $1`, document.GetSaleOrderID()).Scan(&orderCustomerID); err != nil {
		return nil, fmt.Errorf("remittance sale order: %w", err)
	}
	if orderCustomerID != document.GetCustomerID() {
		return nil, errors.New("remittance customer does not match sale order")
	}
	if document.GetSalesInvoiceID() != "" {
		var invoiceCustomerID, invoiceOrderID string
		if err := tx.QueryRow(ctx, `SELECT customer_id::text, sale_order_id::text FROM sales_invoices WHERE id = $1`,
			document.GetSalesInvoiceID()).Scan(&invoiceCustomerID, &invoiceOrderID); err != nil {
			return nil, fmt.Errorf("remittance sales invoice: %w", err)
		}
		if invoiceCustomerID != document.GetCustomerID() || invoiceOrderID != document.GetSaleOrderID() {
			return nil, errors.New("remittance invoice does not match customer and sale order")
		}
	}

	_, err = tx.Exec(ctx, `INSERT INTO remittances
		(id, customer_id, sale_order_id, sales_invoice_id, point_of_sale, document_number, issue_date, delivery_date, status,
		 idempotency_key, created_at, updated_at)
		VALUES ($1,$2,$3,NULLIF($4,'')::uuid,$5,$6,$7,NULLIF($8,'')::date,'DRAFT',$9,$10,$11)`,
		document.GetID(), document.GetCustomerID(), document.GetSaleOrderID(), document.GetSalesInvoiceID(),
		document.GetPointOfSale(), document.GetDocumentNumber(), document.GetIssueDate(), document.GetDeliveryDate(),
		document.GetIdempotencyKey(), document.GetCreatedAt(), document.GetUpdatedAt())
	if err != nil {
		return nil, err
	}
	for _, item := range document.GetItems() {
		quantity, err := valueObject.DecimalToUint64(item.Quantity)
		if err != nil {
			return nil, err
		}
		var orderedQuantity uint64
		if err := tx.QueryRow(ctx, `SELECT quantity FROM sale_order_items WHERE sale_order_id = $1 AND product_id = $2`,
			document.GetSaleOrderID(), item.ProductID).Scan(&orderedQuantity); err != nil {
			return nil, errors.New("remittance item is not part of sale order")
		}
		if quantity > orderedQuantity {
			return nil, errors.New("remittance item quantity exceeds sale order quantity")
		}
		_, err = tx.Exec(ctx, `INSERT INTO remittance_items
			(remittance_id, line_number, product_id, description, quantity, lot_number)
			VALUES ($1,$2,$3,$4,$5,NULLIF($6,''))`,
			document.GetID(), item.LineNumber, item.ProductID, item.Description, item.Quantity, item.LotNumber)
		if err != nil {
			return nil, err
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return r.GetRemittanceByID(ctx, document.GetID())
}

func (r *Repository) GetRemittanceByID(ctx context.Context, id string) (*remittance.Remittance, error) {
	return r.loadRemittance(ctx, id, "")
}

func (r *Repository) ListRemittances(ctx context.Context, filter sales_administration.RemittanceFilter) ([]remittance.Remittance, error) {
	where, args := buildAdministrativeFilter(filter.CustomerID, filter.SaleOrderID, remittanceStorageStatus(filter.Status), "customer_id", "sale_order_id", "status")
	rows, err := r.pool.Query(ctx, `SELECT id::text FROM remittances`+where+` ORDER BY created_at DESC`, args...)
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
	result := make([]remittance.Remittance, 0, len(ids))
	for _, id := range ids {
		document, err := r.GetRemittanceByID(ctx, id)
		if err != nil {
			return nil, err
		}
		result = append(result, *document)
	}
	return result, nil
}

func remittanceStorageStatus(status string) string {
	if status == string(remittance.StatusConfirmed) {
		return "DELIVERED"
	}
	return status
}

func (r *Repository) getRemittanceByIdempotencyKey(ctx context.Context, key string) (*remittance.Remittance, error) {
	return r.loadRemittance(ctx, "", key)
}

// ConfirmRemittance atomically locks the remittance, inserts each physical
// dispatch movement once using the remittance ID as reference, and then marks
// the document with the existing schema's terminal DELIVERED status.
func (r *Repository) ConfirmRemittance(ctx context.Context, id, userID string) (*remittance.Remittance, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	var storedStatus, saleOrderID string
	if err := tx.QueryRow(ctx, `SELECT status, sale_order_id::text FROM remittances WHERE id = $1 FOR UPDATE`, id).Scan(&storedStatus, &saleOrderID); err != nil {
		return nil, err
	}
	if storedStatus == "DELIVERED" {
		if err := tx.Commit(ctx); err != nil {
			return nil, err
		}
		return r.GetRemittanceByID(ctx, id)
	}
	if storedStatus != "DRAFT" && storedStatus != "ISSUED" {
		return nil, remittance.ErrRemittanceImmutable
	}
	// Serialize confirmations for the same sale order so concurrent partial
	// remittances cannot over-dispatch it.
	var lockedOrderID string
	if err := tx.QueryRow(ctx, `SELECT id::text FROM sale_orders WHERE id = $1 FOR UPDATE`, saleOrderID).Scan(&lockedOrderID); err != nil {
		return nil, err
	}

	rows, err := tx.Query(ctx, `SELECT product_id::text, quantity::text, COALESCE(lot_number,'')
		FROM remittance_items WHERE remittance_id = $1 ORDER BY line_number`, id)
	if err != nil {
		return nil, err
	}
	type dispatchItem struct {
		productID, quantity, lotNumber string
	}
	var items []dispatchItem
	for rows.Next() {
		var item dispatchItem
		if err := rows.Scan(&item.productID, &item.quantity, &item.lotNumber); err != nil {
			rows.Close()
			return nil, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return nil, err
	}
	rows.Close()

	for _, item := range items {
		quantity, err := valueObject.DecimalToUint64(item.quantity)
		if err != nil {
			return nil, fmt.Errorf("remittance item quantity cannot be dispatched: %w", err)
		}
		var orderedQuantity uint64
		if err := tx.QueryRow(ctx, `SELECT quantity FROM sale_order_items WHERE sale_order_id = $1 AND product_id = $2`,
			saleOrderID, item.productID).Scan(&orderedQuantity); err != nil {
			return nil, errors.New("remittance item is not part of sale order")
		}
		var alreadyDelivered string
		if err := tx.QueryRow(ctx, `SELECT COALESCE(SUM(ri.quantity),0)::text
			FROM remittance_items ri JOIN remittances rm ON rm.id = ri.remittance_id
			WHERE rm.sale_order_id = $1 AND ri.product_id = $2 AND rm.status = 'DELIVERED'`,
			saleOrderID, item.productID).Scan(&alreadyDelivered); err != nil {
			return nil, err
		}
		deliveredQuantity, err := valueObject.DecimalToUint64(alreadyDelivered)
		if err != nil && alreadyDelivered != "0" {
			return nil, err
		}
		if deliveredQuantity+quantity > orderedQuantity {
			return nil, errors.New("confirmed remittance quantities exceed sale order quantity")
		}
		var inventoryID string
		if err := tx.QueryRow(ctx, `SELECT id::text FROM product_inventories WHERE product_id = $1 ORDER BY id LIMIT 1 FOR UPDATE`,
			item.productID).Scan(&inventoryID); err != nil {
			return nil, fmt.Errorf("remittance product inventory: %w", err)
		}
		_, err = tx.Exec(ctx, `INSERT INTO product_inventory_movements
			(product_inventory_id, movement_type, quantity, stage, reference, lot_number, user_id, created_at)
			SELECT $1::uuid, 'PRODUCT_DISPATCHED', $2::bigint, 'DRESSED', $3::varchar, NULLIF($4,'')::varchar, $5::varchar, $6::varchar
			WHERE NOT EXISTS (
				SELECT 1 FROM product_inventory_movements
				WHERE product_inventory_id = $1::uuid AND movement_type = 'PRODUCT_DISPATCHED' AND reference = $3::varchar
			)`, inventoryID, quantity, id, item.lotNumber, userID, time.Now().UTC().Format(time.RFC3339))
		if err != nil {
			return nil, err
		}
	}
	if _, err := tx.Exec(ctx, `UPDATE remittances SET status = 'DELIVERED', delivery_date = COALESCE(delivery_date, CURRENT_DATE),
		updated_at = NOW() WHERE id = $1`, id); err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return r.GetRemittanceByID(ctx, id)
}

func (r *Repository) loadRemittance(ctx context.Context, id, idempotencyKey string) (*remittance.Remittance, error) {
	query := `SELECT id::text, customer_id::text, sale_order_id::text, COALESCE(sales_invoice_id::text,''), point_of_sale, document_number,
		issue_date::text, COALESCE(delivery_date::text,''), status, idempotency_key, created_at, updated_at
		FROM remittances WHERE id = $1`
	arg := id
	if idempotencyKey != "" {
		query = `SELECT id::text, customer_id::text, sale_order_id::text, COALESCE(sales_invoice_id::text,''), point_of_sale, document_number,
			issue_date::text, COALESCE(delivery_date::text,''), status, idempotency_key, created_at, updated_at
			FROM remittances WHERE idempotency_key = $1`
		arg = idempotencyKey
	}
	var documentID, customerID, saleOrderID, invoiceID, issueDate, deliveryDate, storedStatus, key string
	var pointOfSale int32
	var documentNumber int64
	var createdAt, updatedAt time.Time
	if err := r.pool.QueryRow(ctx, query, arg).Scan(&documentID, &customerID, &saleOrderID, &invoiceID, &pointOfSale, &documentNumber,
		&issueDate, &deliveryDate, &storedStatus, &key, &createdAt, &updatedAt); err != nil {
		return nil, err
	}
	items, err := r.loadRemittanceItems(ctx, documentID)
	if err != nil {
		return nil, err
	}
	status := remittance.StatusDraft
	if storedStatus == "DELIVERED" {
		status = remittance.StatusConfirmed
	}
	document := remittance.ReconstituteRemittance(documentID, remittance.NewRemittanceParams{
		CustomerID: customerID, SaleOrderID: saleOrderID, SalesInvoiceID: invoiceID, PointOfSale: pointOfSale,
		DocumentNumber: documentNumber, IssueDate: issueDate, DeliveryDate: deliveryDate, IdempotencyKey: key, Items: items,
	}, status, createdAt.UTC().Format(time.RFC3339Nano), updatedAt.UTC().Format(time.RFC3339Nano))
	return &document, nil
}

func (r *Repository) loadRemittanceItems(ctx context.Context, remittanceID string) ([]remittance.Item, error) {
	rows, err := r.pool.Query(ctx, `SELECT line_number, COALESCE(product_id::text,''), description, quantity::text,
		COALESCE(lot_number,'') FROM remittance_items WHERE remittance_id = $1 ORDER BY line_number`, remittanceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []remittance.Item
	for rows.Next() {
		var item remittance.Item
		if err := rows.Scan(&item.LineNumber, &item.ProductID, &item.Description, &item.Quantity, &item.LotNumber); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

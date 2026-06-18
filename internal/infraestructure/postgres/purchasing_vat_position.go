package postgres

import (
	"context"

	domain "ricitelli-back/internal/domain/purchasing-administration"
)

func (r *Repository) GetMonthlyVATPosition(ctx context.Context, month string) (*domain.MonthlyVATPosition, error) {
	rows, err := r.pool.Query(ctx, `WITH sales AS (
			SELECT sii.tax_rate,
				ROUND(SUM(CASE WHEN UPPER(si.document_type) LIKE '%CREDIT%'
					THEN -(sii.tax_amount*si.exchange_rate) ELSE sii.tax_amount*si.exchange_rate END),4) sales_debit
			FROM sales_invoice_items sii
			JOIN sales_invoices si ON si.id=sii.sales_invoice_id
			WHERE si.issue_date >= $1::date AND si.issue_date < ($1::date + INTERVAL '1 month')
			  AND si.status NOT IN ('DRAFT','CANCELLED')
			GROUP BY sii.tax_rate
		), purchases AS (
			SELECT sii.tax_rate,
				ROUND(SUM(CASE WHEN UPPER(si.document_type) LIKE '%CREDIT%'
					THEN -(sii.tax_amount*si.exchange_rate) ELSE sii.tax_amount*si.exchange_rate END),4) purchase_credit
			FROM supplier_invoice_items sii
			JOIN supplier_invoices si ON si.id=sii.supplier_invoice_id
			WHERE si.issue_date >= $1::date AND si.issue_date < ($1::date + INTERVAL '1 month')
			  AND si.status NOT IN ('DRAFT','CANCELLED')
			GROUP BY sii.tax_rate
		)
		SELECT COALESCE(s.tax_rate,p.tax_rate)::text,
			COALESCE(s.sales_debit,0)::text,COALESCE(p.purchase_credit,0)::text
		FROM sales s FULL OUTER JOIN purchases p ON p.tax_rate=s.tax_rate
		ORDER BY COALESCE(s.tax_rate,p.tax_rate)`, month+"-01")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var lines []domain.VATPositionLine
	for rows.Next() {
		var line domain.VATPositionLine
		if err := rows.Scan(&line.TaxRate, &line.SalesDebit, &line.PurchaseCredit); err != nil {
			return nil, err
		}
		lines = append(lines, line)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	result, err := domain.NewMonthlyVATPosition(month, lines)
	return &result, err
}

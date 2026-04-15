package reporting

import (
	"context"
	"errors"
	"fmt"

	"ricitelli-back/internal/service/reporting/charts"
	"ricitelli-back/internal/service/reporting/data"
	"ricitelli-back/internal/service/reporting/pdf"
)

// GenerateCustomerReport builds a per-customer commercial summary.
func (s *Service) GenerateCustomerReport(ctx context.Context, customerID, from, to string) (*Result, error) {
	if customerID == "" {
		return nil, errors.New("customer_id is required")
	}
	if from == "" || to == "" {
		return nil, errors.New("from_date and to_date are required")
	}
	agg, err := data.AggregateCustomer(ctx, s.Sales, s.Customers, s.Products, customerID, from, to)
	if err != nil {
		return nil, err
	}

	name := customerID
	group := ""
	if agg.Customer != nil {
		name = agg.Customer.GetSocialReason()
		group = string(agg.Customer.GetGroup())
	}

	r := pdf.New("Informe de Cliente", name)
	r.AddCover(from, to)

	r.Pdf().AddPage()
	r.SectionTitle("Resumen")
	r.AddKPIBoxes([]pdf.KPI{
		{Label: "Órdenes", Value: fmt.Sprintf("%d", agg.OrderCount)},
		{Label: "Ingresos (ARS)", Value: formatCurrency(agg.TotalRevenueARS, "ARS")},
		{Label: "Ticket promedio", Value: formatCurrency(agg.AverageTicket, "ARS")},
		{Label: "Grupo", Value: group},
		{Label: "Primera compra", Value: formatShortDate(agg.FirstPurchase)},
		{Label: "Última compra", Value: formatShortDate(agg.LastPurchase)},
	})

	if agg.OrderCount == 0 {
		r.Paragraph("No se encontraron órdenes para este cliente en el rango seleccionado.")
		return s.saveAndResult(r, TypeCustomer, from, to)
	}

	if png, err := charts.Lines("Ingresos por mes (ARS)", []charts.TimeSeries{{
		Name: "Ingresos", Points: toTimePoints(agg.RevenueByMonth),
	}}); err == nil {
		_ = r.AddChartPNG("cu_rev", png, 180)
	}
	if png, err := charts.HorizontalBar("Productos favoritos", nameValueToBar(agg.TopProducts)); err == nil {
		_ = r.AddChartPNG("cu_prods", png, 180)
	}
	if png, err := charts.Donut("Ingresos por moneda", currencyToBar(agg.RevenueByCurrency)); err == nil {
		_ = r.AddChartPNG("cu_curr", png, 160)
	}

	r.SectionTitle("Órdenes del período")
	rows := make([]pdf.TableRow, 0, len(agg.Orders))
	for _, o := range agg.Orders {
		var amount float64
		for _, it := range o.GetItems() {
			amount += float64(it.Quantity) * float64(it.UnitPrice)
		}
		rows = append(rows, pdf.TableRow{
			formatShortDate(o.GetCreatedAt()),
			string(o.GetStatus()),
			string(o.GetMarket()),
			string(o.GetCurrency()),
			formatNumber(amount),
		})
	}
	r.AddTable(
		[]string{"Fecha", "Estado", "Mercado", "Moneda", "Monto"},
		rows,
		[]float64{25, 30, 30, 25, 40},
	)

	return s.saveAndResult(r, TypeCustomer, from, to)
}

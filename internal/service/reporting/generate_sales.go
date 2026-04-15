package reporting

import (
	"context"
	"errors"
	"fmt"
	"time"

	"ricitelli-back/internal/service/reporting/charts"
	"ricitelli-back/internal/service/reporting/data"
	"ricitelli-back/internal/service/reporting/pdf"
)

// GenerateSalesReport orchestrates aggregation + charts + PDF build for the sales report.
func (s *Service) GenerateSalesReport(ctx context.Context, f ReportFilter) (*Result, error) {
	if err := f.Validate(true); err != nil {
		return nil, err
	}

	agg, err := data.AggregateSales(ctx, s.Sales, s.Customers, s.Products,
		f.FromDate, f.ToDate,
		data.SalesFilter{Market: f.Market, Currency: f.Currency, CustomerID: f.CustomerID, ProductID: f.ProductID})
	if err != nil {
		return nil, err
	}

	r := pdf.New("Informe de Ventas", "")
	r.AddCover(f.FromDate, f.ToDate)

	// Executive summary
	r.Pdf().AddPage()
	r.SectionTitle("Resumen ejecutivo")
	r.AddKPIBoxes([]pdf.KPI{
		{Label: "Ingresos (ARS)", Value: formatCurrency(agg.TotalRevenueARS, "ARS"), Sub: "monto total facturado"},
		{Label: "Órdenes", Value: fmt.Sprintf("%d", agg.OrderCount), Sub: "en el período"},
		{Label: "Unidades despachadas", Value: fmt.Sprintf("%d", agg.UnitsDispatched), Sub: "botellas"},
		{Label: "Ticket promedio (ARS)", Value: formatCurrency(agg.AverageTicket, "ARS"), Sub: "por orden"},
		{Label: "Tasa de cumplimiento", Value: fmt.Sprintf("%.1f%%", agg.FulfillmentRate), Sub: "DISPATCHED / total"},
	})

	if agg.OrderCount == 0 {
		r.Paragraph("No se encontraron órdenes de venta en el rango seleccionado.")
		return s.saveAndResult(r, TypeSales, f.FromDate, f.ToDate)
	}

	// Chart 1: Revenue by day
	if png, err := charts.Lines("Evolución de ingresos (ARS) por día", []charts.TimeSeries{{
		Name:   "Ingresos",
		Points: toTimePoints(agg.RevenueByDay),
	}}); err == nil {
		_ = r.AddChartPNG("rev_day", png, 180)
	}

	// Chart 2: Orders by market (donut)
	if png, err := charts.Donut("Órdenes por mercado", mapToBarData(agg.OrdersByMarket)); err == nil {
		_ = r.AddChartPNG("ord_market", png, 160)
	}

	// Chart 3: Top products by revenue (horizontal bar)
	if png, err := charts.HorizontalBar("Top 10 productos por ingresos (ARS)", nameValueToBar(agg.TopProductsByRevenue)); err == nil {
		_ = r.AddChartPNG("top_prod", png, 180)
	}

	// Chart 4: Revenue by currency (vertical bar)
	if png, err := charts.VerticalBar("Ingresos por moneda", currencyToBar(agg.RevenueByCurrency)); err == nil {
		_ = r.AddChartPNG("rev_curr", png, 160)
	}

	// Chart 5: Sales by customer group
	if png, err := charts.HorizontalBar("Ventas por grupo de cliente", mapToBarData(agg.SalesByCustomerGroup)); err == nil {
		_ = r.AddChartPNG("sales_group", png, 180)
	}

	// Chart 6: Orders by country
	if png, err := charts.HorizontalBar("Órdenes por país destino", mapToBarData(agg.OrdersByCountry)); err == nil {
		_ = r.AddChartPNG("ord_country", png, 180)
	}

	// Chart 7: Status distribution
	if png, err := charts.HorizontalBar("Distribución por estado", mapToBarData(agg.StatusDistribution)); err == nil {
		_ = r.AddChartPNG("status", png, 180)
	}

	// Top orders table
	r.Pdf().AddPage()
	r.SectionTitle("Top 20 órdenes por monto")
	rows := make([]pdf.TableRow, 0, len(agg.TopOrders))
	for _, o := range agg.TopOrders {
		rows = append(rows, pdf.TableRow{
			formatShortDate(o.CreatedAt),
			o.CustomerName,
			o.ProductName,
			formatCurrency(o.Amount, o.Currency),
			o.Currency,
			o.Status,
		})
	}
	r.AddTable(
		[]string{"Fecha", "Cliente", "Producto", "Monto", "Moneda", "Estado"},
		rows,
		[]float64{22, 50, 40, 28, 15, 25},
	)

	return s.saveAndResult(r, TypeSales, f.FromDate, f.ToDate)
}

func (s *Service) saveAndResult(r *pdf.Report, t ReportType, from, to string) (*Result, error) {
	bs, err := r.Build()
	if err != nil {
		return nil, err
	}
	if s.Storage == nil {
		return nil, errors.New("storage not configured")
	}
	m, err := s.Storage.Save(bs, string(t), from, to)
	if err != nil {
		return nil, err
	}
	return s.buildResult(m), nil
}

func toTimePoints(days []data.DayValue) []charts.TimePoint {
	out := make([]charts.TimePoint, len(days))
	for i, d := range days {
		out[i] = charts.TimePoint{Time: d.Day, Value: d.Value}
	}
	return out
}

func mapToBarData(m map[string]int) []charts.BarDatum {
	out := make([]charts.BarDatum, 0, len(m))
	for k, v := range m {
		out = append(out, charts.BarDatum{Label: k, Value: float64(v)})
	}
	// Sort by value desc
	for i := 0; i < len(out); i++ {
		for j := i + 1; j < len(out); j++ {
			if out[j].Value > out[i].Value {
				out[i], out[j] = out[j], out[i]
			}
		}
	}
	return out
}

func currencyToBar(m map[string]float64) []charts.BarDatum {
	out := make([]charts.BarDatum, 0, len(m))
	for k, v := range m {
		out = append(out, charts.BarDatum{Label: k, Value: v})
	}
	return out
}

func nameValueToBar(list []data.NameValue) []charts.BarDatum {
	out := make([]charts.BarDatum, len(list))
	for i, d := range list {
		out[i] = charts.BarDatum{Label: d.Name, Value: d.Value}
	}
	return out
}

func formatCurrency(v float64, currency string) string {
	return fmt.Sprintf("%s %s", currency, formatNumber(v))
}

func formatNumber(v float64) string {
	s := fmt.Sprintf("%.2f", v)
	// Insert thousands separator (dots) — simple implementation
	// Find dot
	intPart := s
	fracPart := ""
	for i, c := range s {
		if c == '.' {
			intPart = s[:i]
			fracPart = s[i:]
			break
		}
	}
	// Insert dots every 3 digits from the right
	n := len(intPart)
	if n <= 3 {
		return intPart + fracPart
	}
	first := n % 3
	out := ""
	if first > 0 {
		out = intPart[:first]
	}
	for i := first; i < n; i += 3 {
		if out != "" {
			out += "."
		}
		out += intPart[i : i+3]
	}
	return out + fracPart
}

func formatShortDate(rfc string) string {
	t, err := time.Parse(time.RFC3339, rfc)
	if err != nil {
		return rfc
	}
	return t.Format("02/01/06")
}

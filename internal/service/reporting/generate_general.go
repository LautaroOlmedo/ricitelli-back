package reporting

import (
	"context"
	"fmt"

	"ricitelli-back/internal/service/reporting/charts"
	"ricitelli-back/internal/service/reporting/data"
	"ricitelli-back/internal/service/reporting/pdf"
)

// GenerateGeneralReport builds an executive summary combining sales + production + inventory health.
func (s *Service) GenerateGeneralReport(ctx context.Context, f ReportFilter) (*Result, error) {
	if err := f.Validate(true); err != nil {
		return nil, err
	}

	salesAgg, err := data.AggregateSales(ctx, s.Sales, s.Customers, s.Products, f.FromDate, f.ToDate, data.SalesFilter{})
	if err != nil {
		return nil, err
	}
	prodAgg, err := data.AggregateProduction(ctx, s.Inventory, s.Production, f.FromDate, f.ToDate)
	if err != nil {
		return nil, err
	}
	invReport, err := s.Inventory.GetInventoryReport(ctx)
	if err != nil {
		return nil, err
	}

	r := pdf.New("Informe General", "")
	r.AddCover(f.FromDate, f.ToDate)

	// Page 2: Executive dashboard (12 KPIs)
	r.Pdf().AddPage()
	r.SectionTitle("Dashboard ejecutivo")
	lowCount := 0
	for _, a := range invReport.DrySupplyAlerts {
		if a.IsLow {
			lowCount++
		}
	}
	r.AddKPIBoxes([]pdf.KPI{
		{Label: "Ingresos (ARS)", Value: formatCurrency(salesAgg.TotalRevenueARS, "ARS")},
		{Label: "Órdenes", Value: fmt.Sprintf("%d", salesAgg.OrderCount)},
		{Label: "Unidades despachadas", Value: fmt.Sprintf("%d", salesAgg.UnitsDispatched)},
		{Label: "Cumplimiento", Value: fmt.Sprintf("%.1f%%", salesAgg.FulfillmentRate)},
		{Label: "SV producidas", Value: fmt.Sprintf("%d", prodAgg.SVProduced)},
		{Label: "Conversiones SV→PT", Value: fmt.Sprintf("%d", prodAgg.SVtoPTConverted)},
		{Label: "Insumos consumidos", Value: fmt.Sprintf("%d", prodAgg.SuppliesConsumed)},
		{Label: "OP completadas", Value: fmt.Sprintf("%d", prodAgg.POCompleted)},
		{Label: "Productos activos", Value: fmt.Sprintf("%d", len(invReport.Products))},
		{Label: "Insumos activos", Value: fmt.Sprintf("%d", len(invReport.DrySupplyAlerts))},
		{Label: "Alertas stock bajo", Value: fmt.Sprintf("%d", lowCount)},
		{Label: "Ticket promedio", Value: formatCurrency(salesAgg.AverageTicket, "ARS")},
	})

	// Page 3: Revenue + conversions
	r.Pdf().AddPage()
	r.SectionTitle("Tendencias")
	if png, err := charts.Lines("Ingresos (ARS) por día", []charts.TimeSeries{{
		Name: "Ingresos", Points: toTimePoints(salesAgg.RevenueByDay),
	}}); err == nil {
		_ = r.AddChartPNG("gen_rev", png, 180)
	}
	if png, err := charts.Lines("Conversiones SV→PT por día", []charts.TimeSeries{{
		Name: "Conversiones", Points: toTimePoints(prodAgg.ConversionsByDay),
	}}); err == nil {
		_ = r.AddChartPNG("gen_conv", png, 180)
	}

	// Page 4: Inventory health
	r.Pdf().AddPage()
	r.SectionTitle("Salud de inventario")

	// Product tricapa (stacked bar)
	if len(invReport.Products) > 0 {
		physical := make([]charts.BarDatum, 0)
		committed := make([]charts.BarDatum, 0)
		available := make([]charts.BarDatum, 0)
		for _, p := range invReport.Products {
			physical = append(physical, charts.BarDatum{Label: p.ProductName, Value: float64(p.DressedPhysical)})
			committed = append(committed, charts.BarDatum{Label: p.ProductName, Value: float64(p.DressedCommitted)})
			available = append(available, charts.BarDatum{Label: p.ProductName, Value: float64(p.DressedAvailable)})
		}
		series := []charts.StackedSeries{
			{Name: "Disponible", Data: available},
			{Name: "Comprometido", Data: committed},
			{Name: "Físico", Data: physical},
		}
		if png, err := charts.StackedBar("Stock PT por producto", series); err == nil {
			_ = r.AddChartPNG("prod_tri", png, 180)
		}
	}

	// Low stock table
	if lowCount > 0 {
		rows := make([]pdf.TableRow, 0, lowCount)
		for _, a := range invReport.DrySupplyAlerts {
			if !a.IsLow {
				continue
			}
			rows = append(rows, pdf.TableRow{
				a.Code,
				a.Name,
				fmt.Sprintf("%d", a.Physical),
				fmt.Sprintf("%d", a.Committed),
				fmt.Sprintf("%d", a.Available),
			})
		}
		r.SectionTitle("Alertas de stock bajo")
		r.AddTable(
			[]string{"Código", "Nombre", "Físico", "Comprometido", "Disponible"},
			rows,
			[]float64{28, 80, 22, 28, 22},
		)
	}

	// Pages 5-6: Sales summary
	r.Pdf().AddPage()
	r.SectionTitle("Resumen de ventas")
	if png, err := charts.Donut("Órdenes por mercado", mapToBarData(salesAgg.OrdersByMarket)); err == nil {
		_ = r.AddChartPNG("g_market", png, 160)
	}
	if png, err := charts.HorizontalBar("Top 10 productos por ingresos", nameValueToBar(salesAgg.TopProductsByRevenue)); err == nil {
		_ = r.AddChartPNG("g_top", png, 180)
	}
	if png, err := charts.HorizontalBar("Distribución por estado", mapToBarData(salesAgg.StatusDistribution)); err == nil {
		_ = r.AddChartPNG("g_status", png, 180)
	}

	// Pages 7-8: Production summary
	r.Pdf().AddPage()
	r.SectionTitle("Resumen de producción")
	if len(prodAgg.SVvsPTByProduct) > 0 {
		series := []charts.StackedSeries{
			{Name: "SV", Data: stackedToBarSV2(prodAgg.SVvsPTByProduct)},
			{Name: "PT", Data: stackedToBarPT2(prodAgg.SVvsPTByProduct)},
		}
		if png, err := charts.StackedBar("SV vs PT por producto", series); err == nil {
			_ = r.AddChartPNG("g_svpt", png, 180)
		}
	}
	if png, err := charts.HorizontalBar("Top insumos consumidos", nameValueToBar(prodAgg.TopConsumedSupplies)); err == nil {
		_ = r.AddChartPNG("g_top_sup", png, 180)
	}

	return s.saveAndResult(r, TypeGeneral, f.FromDate, f.ToDate)
}

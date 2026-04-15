package reporting

import (
	"context"
	"fmt"

	"ricitelli-back/internal/service/reporting/charts"
	"ricitelli-back/internal/service/reporting/data"
	"ricitelli-back/internal/service/reporting/pdf"
)

// GenerateProductionReport builds the production PDF report.
func (s *Service) GenerateProductionReport(ctx context.Context, f ReportFilter) (*Result, error) {
	if err := f.Validate(true); err != nil {
		return nil, err
	}
	agg, err := data.AggregateProduction(ctx, s.Inventory, s.Production, f.FromDate, f.ToDate)
	if err != nil {
		return nil, err
	}

	r := pdf.New("Informe de Producción", "")
	r.AddCover(f.FromDate, f.ToDate)

	r.Pdf().AddPage()
	r.SectionTitle("Resumen ejecutivo")
	r.AddKPIBoxes([]pdf.KPI{
		{Label: "SV producidas", Value: fmt.Sprintf("%d", agg.SVProduced), Sub: "botellas sin vestir"},
		{Label: "Conversiones SV→PT", Value: fmt.Sprintf("%d", agg.SVtoPTConverted), Sub: "botellas vestidas"},
		{Label: "Despachadas", Value: fmt.Sprintf("%d", agg.Dispatched), Sub: "salidas del depósito"},
		{Label: "Insumos ingresados", Value: fmt.Sprintf("%d", agg.SuppliesIn), Sub: "unidades"},
		{Label: "Insumos consumidos", Value: fmt.Sprintf("%d", agg.SuppliesConsumed), Sub: "unidades"},
		{Label: "OP completadas", Value: fmt.Sprintf("%d", agg.POCompleted), Sub: "órdenes de producción"},
		{Label: "OP en curso", Value: fmt.Sprintf("%d", agg.POInProgress), Sub: ""},
		{Label: "OP canceladas", Value: fmt.Sprintf("%d", agg.POCancelled), Sub: ""},
	})

	// Chart 1: SV vs PT per product (stacked bar)
	if len(agg.SVvsPTByProduct) > 0 {
		series := []charts.StackedSeries{
			{Name: "SV", Data: stackedToBarSV2(agg.SVvsPTByProduct)},
			{Name: "PT", Data: stackedToBarPT2(agg.SVvsPTByProduct)},
		}
		if png, err := charts.StackedBar("SV vs PT por producto", series); err == nil {
			_ = r.AddChartPNG("sv_pt", png, 180)
		}
	}

	// Chart 2: conversions by day
	if png, err := charts.Lines("Conversiones SV→PT por día", []charts.TimeSeries{{
		Name: "Conversiones", Points: toTimePoints(agg.ConversionsByDay),
	}}); err == nil {
		_ = r.AddChartPNG("conv_day", png, 180)
	}

	// Chart 3: supplies in vs consumed by day
	if len(agg.SuppliesInVsConsByDay) > 0 {
		in := make([]charts.TimePoint, len(agg.SuppliesInVsConsByDay))
		out := make([]charts.TimePoint, len(agg.SuppliesInVsConsByDay))
		for i, p := range agg.SuppliesInVsConsByDay {
			in[i] = charts.TimePoint{Time: p.Day, Value: p.A}
			out[i] = charts.TimePoint{Time: p.Day, Value: p.B}
		}
		if png, err := charts.Lines("Ingresos vs consumos de insumos por día", []charts.TimeSeries{
			{Name: "Ingresados", Points: in},
			{Name: "Consumidos", Points: out},
		}); err == nil {
			_ = r.AddChartPNG("supplies_io", png, 180)
		}
	}

	// Chart 4: top consumed supplies
	if png, err := charts.HorizontalBar("Top 10 insumos consumidos", nameValueToBar(agg.TopConsumedSupplies)); err == nil {
		_ = r.AddChartPNG("top_consumed", png, 180)
	}

	// Chart 5: lots by day
	if png, err := charts.Lines("Lotes generados por día", []charts.TimeSeries{{
		Name: "Lotes", Points: toTimePoints(agg.LotsByDay),
	}}); err == nil {
		_ = r.AddChartPNG("lots_day", png, 180)
	}

	// Chart 6: PO status
	if png, err := charts.Donut("Estado de órdenes de producción", mapToBarData(agg.POStatus)); err == nil {
		_ = r.AddChartPNG("po_status", png, 160)
	}

	// Recent movements table
	r.Pdf().AddPage()
	r.SectionTitle("Movimientos recientes")
	rows := make([]pdf.TableRow, 0, len(agg.RecentMovements))
	for _, m := range agg.RecentMovements {
		rows = append(rows, pdf.TableRow{
			formatShortDate(m.CreatedAt),
			m.UserID,
			m.MovementType,
			m.ItemName,
			fmt.Sprintf("%d", m.Quantity),
			m.LotNumber,
		})
	}
	r.AddTable(
		[]string{"Fecha", "Usuario", "Tipo", "Ítem", "Cantidad", "Lote"},
		rows,
		[]float64{22, 30, 40, 45, 20, 25},
	)

	return s.saveAndResult(r, TypeProduction, f.FromDate, f.ToDate)
}

// Helper conversions (internal use in this file)
func stackedToBarSV2(items []data.StackedProduct) []charts.BarDatum {
	out := make([]charts.BarDatum, len(items))
	for i, p := range items {
		out[i] = charts.BarDatum{Label: p.ProductName, Value: p.SV}
	}
	return out
}

func stackedToBarPT2(items []data.StackedProduct) []charts.BarDatum {
	out := make([]charts.BarDatum, len(items))
	for i, p := range items {
		out[i] = charts.BarDatum{Label: p.ProductName, Value: p.PT}
	}
	return out
}

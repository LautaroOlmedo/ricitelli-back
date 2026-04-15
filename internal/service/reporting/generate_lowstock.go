package reporting

import (
	"context"
	"fmt"

	"ricitelli-back/internal/service/reporting/charts"
	"ricitelli-back/internal/service/reporting/data"
	"ricitelli-back/internal/service/reporting/pdf"
)

// GenerateLowStockReport builds a report with items below reorder_point and reorder suggestions.
func (s *Service) GenerateLowStockReport(ctx context.Context) (*Result, error) {
	items, err := data.AggregateLowStock(ctx, s.Inventory)
	if err != nil {
		return nil, err
	}

	r := pdf.New("Informe de Stock Bajo", "Insumos que requieren reposición")
	r.AddCover("", "")

	r.Pdf().AddPage()
	r.SectionTitle("Resumen")
	r.AddKPIBoxes([]pdf.KPI{
		{Label: "Insumos críticos", Value: fmt.Sprintf("%d", len(items)), Sub: "por debajo del umbral"},
	})

	if len(items) == 0 {
		r.Paragraph("No se detectaron insumos con stock bajo.")
		return s.saveAndResult(r, TypeLowStock, "", "")
	}

	// Chart: top 15 by criticality
	topItems := items
	if len(topItems) > 15 {
		topItems = topItems[:15]
	}
	data := make([]charts.BarDatum, len(topItems))
	for i, it := range topItems {
		data[i] = charts.BarDatum{Label: it.Name, Value: it.Criticality * 100}
	}
	if png, err := charts.HorizontalBar("Criticidad (% bajo el umbral)", data); err == nil {
		_ = r.AddChartPNG("low_crit", png, 180)
	}

	// Table
	r.SectionTitle("Detalle de insumos críticos")
	rows := make([]pdf.TableRow, 0, len(items))
	for _, it := range items {
		rows = append(rows, pdf.TableRow{
			it.Code,
			it.Name,
			fmt.Sprintf("%d", it.Physical),
			fmt.Sprintf("%d", it.Committed),
			fmt.Sprintf("%d", it.Available),
			fmt.Sprintf("%.1f", it.AvgDailyConsumption),
			fmt.Sprintf("%d", it.RecommendedReorder),
		})
	}
	r.AddTable(
		[]string{"Código", "Nombre", "Físico", "Comprometido", "Disponible", "Consumo/día", "Reponer"},
		rows,
		[]float64{22, 60, 18, 25, 20, 20, 20},
	)

	return s.saveAndResult(r, TypeLowStock, "", "")
}

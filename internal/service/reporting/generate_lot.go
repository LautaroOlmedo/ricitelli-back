package reporting

import (
	"context"
	"errors"
	"fmt"

	"ricitelli-back/internal/service/reporting/data"
	"ricitelli-back/internal/service/reporting/pdf"
)

// GenerateLotTraceabilityReport reconstructs the chain of movements for a single lot_number.
func (s *Service) GenerateLotTraceabilityReport(ctx context.Context, lotNumber string) (*Result, error) {
	if lotNumber == "" {
		return nil, errors.New("lot_number is required")
	}
	trace, err := data.AggregateLotTrace(ctx, s.Inventory, lotNumber)
	if err != nil {
		return nil, err
	}

	r := pdf.New("Trazabilidad de Lote", lotNumber)
	r.AddCover("", "")

	r.Pdf().AddPage()
	r.SectionTitle("Información del lote")
	r.AddKPIBoxes([]pdf.KPI{
		{Label: "Lote", Value: trace.LotNumber},
		{Label: "Producto", Value: trace.ProductName},
		{Label: "Unidades", Value: fmt.Sprintf("%d", trace.TotalUnits)},
		{Label: "Movimientos", Value: fmt.Sprintf("%d", len(trace.Movements))},
	})

	if len(trace.Movements) == 0 {
		r.Paragraph(fmt.Sprintf("No se encontraron movimientos para el lote %s.", lotNumber))
		return s.saveAndResult(r, TypeLot, "", "")
	}

	r.SectionTitle("Cadena de movimientos")
	rows := make([]pdf.TableRow, 0, len(trace.Movements))
	for _, m := range trace.Movements {
		rows = append(rows, pdf.TableRow{
			formatShortDate(m.CreatedAt),
			m.UserID,
			m.MovementType,
			m.Stage,
			fmt.Sprintf("%d", m.Quantity),
			m.Reference,
		})
	}
	r.AddTable(
		[]string{"Fecha", "Usuario", "Tipo", "Etapa", "Cantidad", "Referencia"},
		rows,
		[]float64{22, 30, 40, 20, 20, 48},
	)

	return s.saveAndResult(r, TypeLot, "", "")
}

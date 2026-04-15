// Package pdf builds PDF reports with a simple API on top of gofpdf.
package pdf

import (
	"bytes"
	"fmt"
	"image"
	_ "image/png"
	"time"

	"github.com/jung-kurt/gofpdf"
)

// Report is a wrapper around gofpdf.Fpdf with helpers for covers, KPIs, images and tables.
type Report struct {
	pdf         *gofpdf.Fpdf
	title       string
	subtitle    string
	generatedAt time.Time
}

// New creates a new A4 report.
func New(title, subtitle string) *Report {
	p := gofpdf.New("P", "mm", "A4", "")
	p.SetMargins(15, 20, 15)
	p.SetAutoPageBreak(true, 20)
	p.SetFont("Helvetica", "", 10)
	return &Report{pdf: p, title: title, subtitle: subtitle, generatedAt: time.Now().UTC()}
}

// AddCover renders a styled cover page.
func (r *Report) AddCover(fromDate, toDate string) {
	r.pdf.AddPage()
	// Brand color bar
	r.pdf.SetFillColor(78, 7, 7)
	r.pdf.Rect(0, 0, 210, 6, "F")

	r.pdf.SetY(70)
	r.pdf.SetFont("Helvetica", "B", 28)
	r.pdf.SetTextColor(40, 40, 40)
	r.pdf.CellFormat(0, 14, r.title, "", 1, "C", false, 0, "")

	r.pdf.SetY(90)
	r.pdf.SetFont("Helvetica", "", 14)
	r.pdf.SetTextColor(100, 100, 100)
	r.pdf.CellFormat(0, 8, "Bodega Ricitelli", "", 1, "C", false, 0, "")

	if r.subtitle != "" {
		r.pdf.Ln(4)
		r.pdf.SetFont("Helvetica", "I", 12)
		r.pdf.CellFormat(0, 6, r.subtitle, "", 1, "C", false, 0, "")
	}

	if fromDate != "" && toDate != "" {
		r.pdf.Ln(10)
		r.pdf.SetFont("Helvetica", "", 11)
		from := formatDate(fromDate)
		to := formatDate(toDate)
		r.pdf.CellFormat(0, 6, fmt.Sprintf("Período: %s  —  %s", from, to), "", 1, "C", false, 0, "")
	}

	r.pdf.SetY(240)
	r.pdf.SetFont("Helvetica", "I", 9)
	r.pdf.SetTextColor(150, 150, 150)
	r.pdf.CellFormat(0, 5, fmt.Sprintf("Generado: %s", r.generatedAt.Format("2006-01-02 15:04 UTC")), "", 1, "C", false, 0, "")

	// Reset
	r.pdf.SetTextColor(40, 40, 40)
}

// SectionTitle adds a large section heading on the current page, adding a new page if needed.
func (r *Report) SectionTitle(text string) {
	if r.pdf.GetY() > 250 {
		r.pdf.AddPage()
	}
	r.pdf.Ln(4)
	r.pdf.SetFont("Helvetica", "B", 14)
	r.pdf.SetTextColor(78, 7, 7)
	r.pdf.CellFormat(0, 8, text, "", 1, "L", false, 0, "")
	r.pdf.SetTextColor(40, 40, 40)
	r.pdf.SetDrawColor(200, 200, 200)
	r.pdf.Line(15, r.pdf.GetY(), 195, r.pdf.GetY())
	r.pdf.Ln(3)
	r.pdf.SetFont("Helvetica", "", 10)
}

// KPI holds a single metric box.
type KPI struct {
	Label string
	Value string
	Sub   string
}

// AddKPIBoxes renders KPIs in a grid (columns auto).
func (r *Report) AddKPIBoxes(kpis []KPI) {
	if len(kpis) == 0 {
		return
	}
	cols := 4
	if len(kpis) < 4 {
		cols = len(kpis)
	}
	cellW := 180.0 / float64(cols)
	cellH := 22.0

	startX := r.pdf.GetX()
	startY := r.pdf.GetY()
	col := 0
	row := 0
	for _, k := range kpis {
		x := startX + float64(col)*cellW
		y := startY + float64(row)*(cellH+2)
		r.pdf.SetFillColor(250, 247, 242)
		r.pdf.SetDrawColor(220, 210, 195)
		r.pdf.Rect(x, y, cellW-2, cellH, "FD")

		r.pdf.SetXY(x+2, y+2)
		r.pdf.SetFont("Helvetica", "", 8)
		r.pdf.SetTextColor(120, 120, 120)
		r.pdf.CellFormat(cellW-4, 4, k.Label, "", 0, "L", false, 0, "")

		r.pdf.SetXY(x+2, y+7)
		r.pdf.SetFont("Helvetica", "B", 14)
		r.pdf.SetTextColor(78, 7, 7)
		r.pdf.CellFormat(cellW-4, 8, k.Value, "", 0, "L", false, 0, "")

		if k.Sub != "" {
			r.pdf.SetXY(x+2, y+15)
			r.pdf.SetFont("Helvetica", "", 7)
			r.pdf.SetTextColor(140, 140, 140)
			r.pdf.CellFormat(cellW-4, 4, k.Sub, "", 0, "L", false, 0, "")
		}

		col++
		if col >= cols {
			col = 0
			row++
		}
	}
	totalRows := row
	if col > 0 {
		totalRows++
	}
	r.pdf.SetY(startY + float64(totalRows)*(cellH+2) + 4)
	r.pdf.SetTextColor(40, 40, 40)
}

// AddChartPNG embeds a PNG chart. width in mm; height auto-calculated from aspect ratio.
func (r *Report) AddChartPNG(name string, pngBytes []byte, widthMM float64) error {
	// Decode to get dimensions for aspect ratio
	img, _, err := image.Decode(bytes.NewReader(pngBytes))
	if err != nil {
		return err
	}
	b := img.Bounds()
	aspect := float64(b.Dy()) / float64(b.Dx())
	heightMM := widthMM * aspect

	// If it doesn't fit on this page, add one.
	if r.pdf.GetY()+heightMM > 275 {
		r.pdf.AddPage()
	}

	// Register and draw
	opt := gofpdf.ImageOptions{ImageType: "PNG", ReadDpi: false}
	r.pdf.RegisterImageOptionsReader(name, opt, bytes.NewReader(pngBytes))
	x := (210 - widthMM) / 2
	r.pdf.ImageOptions(name, x, r.pdf.GetY(), widthMM, heightMM, true, opt, 0, "")
	r.pdf.Ln(3)
	return nil
}

// TableRow is a generic row of string cells.
type TableRow []string

// AddTable renders a simple table with headers.
func (r *Report) AddTable(headers []string, rows []TableRow, widths []float64) {
	if len(widths) != len(headers) {
		// Equal widths if mismatch
		widths = make([]float64, len(headers))
		each := 180.0 / float64(len(headers))
		for i := range widths {
			widths[i] = each
		}
	}

	// Header
	r.pdf.SetFont("Helvetica", "B", 9)
	r.pdf.SetFillColor(78, 7, 7)
	r.pdf.SetTextColor(255, 255, 255)
	for i, h := range headers {
		r.pdf.CellFormat(widths[i], 7, h, "1", 0, "L", true, 0, "")
	}
	r.pdf.Ln(-1)

	// Rows
	r.pdf.SetFont("Helvetica", "", 8)
	r.pdf.SetTextColor(40, 40, 40)
	alt := false
	for _, row := range rows {
		if r.pdf.GetY() > 270 {
			r.pdf.AddPage()
			// Repeat header
			r.pdf.SetFont("Helvetica", "B", 9)
			r.pdf.SetFillColor(78, 7, 7)
			r.pdf.SetTextColor(255, 255, 255)
			for i, h := range headers {
				r.pdf.CellFormat(widths[i], 7, h, "1", 0, "L", true, 0, "")
			}
			r.pdf.Ln(-1)
			r.pdf.SetFont("Helvetica", "", 8)
			r.pdf.SetTextColor(40, 40, 40)
		}
		if alt {
			r.pdf.SetFillColor(248, 245, 240)
		} else {
			r.pdf.SetFillColor(255, 255, 255)
		}
		for i, cell := range row {
			if i >= len(widths) {
				break
			}
			r.pdf.CellFormat(widths[i], 6, truncateStr(cell, int(widths[i]/1.8)), "1", 0, "L", true, 0, "")
		}
		r.pdf.Ln(-1)
		alt = !alt
	}
	r.pdf.Ln(3)
}

// Paragraph adds a paragraph of text.
func (r *Report) Paragraph(text string) {
	r.pdf.SetFont("Helvetica", "", 10)
	r.pdf.SetTextColor(60, 60, 60)
	r.pdf.MultiCell(0, 5, text, "", "L", false)
	r.pdf.Ln(2)
}

// AddFooter renders a footer on every page (call after content is done if needed).
func (r *Report) AddFooter() {
	// gofpdf supports SetFooterFunc, but we set it early for all pages.
	// Instead we register a footer in Build and re-render. Left for future extension.
}

// Build finalizes the PDF and returns bytes. It also registers a footer on all pages.
func (r *Report) Build() ([]byte, error) {
	// Footer function for all pages
	r.pdf.SetFooterFunc(func() {
		r.pdf.SetY(-15)
		r.pdf.SetFont("Helvetica", "I", 7)
		r.pdf.SetTextColor(150, 150, 150)
		r.pdf.CellFormat(0, 4, fmt.Sprintf("Bodega Ricitelli · Página %d/{nb} · Generado %s",
			r.pdf.PageNo(), r.generatedAt.Format("2006-01-02 15:04 UTC")),
			"", 0, "C", false, 0, "")
	})
	r.pdf.AliasNbPages("")

	var buf bytes.Buffer
	if err := r.pdf.Output(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// Pdf exposes the underlying gofpdf instance for advanced layouts.
func (r *Report) Pdf() *gofpdf.Fpdf { return r.pdf }

func formatDate(rfc3339 string) string {
	t, err := time.Parse(time.RFC3339, rfc3339)
	if err != nil {
		return rfc3339
	}
	return t.Format("02/01/2006")
}

func truncateStr(s string, max int) string {
	if max <= 0 || len(s) <= max {
		return s
	}
	if max <= 3 {
		return s[:max]
	}
	return s[:max-1] + "…"
}

package charts

import (
	"bytes"

	"github.com/wcharczuk/go-chart/v2"
	"github.com/wcharczuk/go-chart/v2/drawing"
)

// Donut renders a donut chart (pie with hole) to PNG.
func Donut(title string, data []BarDatum) ([]byte, error) {
	if len(data) == 0 || sumValues(data) == 0 {
		return placeholder(title, "Sin datos")
	}
	values := make([]chart.Value, len(data))
	for i, d := range data {
		c := ColorAt(i)
		values[i] = chart.Value{
			Label: d.Label,
			Value: d.Value,
			Style: chart.Style{
				FillColor: drawing.Color{R: c.R, G: c.G, B: c.B, A: c.A},
			},
		}
	}
	ch := chart.DonutChart{
		Title:      title,
		TitleStyle: chart.Style{FontSize: 14},
		Background: chart.Style{Padding: chart.Box{Top: 40, Left: 20, Right: 20, Bottom: 20}},
		Height:     DefaultHeight,
		Width:      DefaultWidth,
		Values:     values,
	}
	var buf bytes.Buffer
	if err := ch.Render(chart.PNG, &buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// Pie renders a pie chart to PNG.
func Pie(title string, data []BarDatum) ([]byte, error) {
	if len(data) == 0 || sumValues(data) == 0 {
		return placeholder(title, "Sin datos")
	}
	values := make([]chart.Value, len(data))
	for i, d := range data {
		c := ColorAt(i)
		values[i] = chart.Value{
			Label: d.Label,
			Value: d.Value,
			Style: chart.Style{
				FillColor: drawing.Color{R: c.R, G: c.G, B: c.B, A: c.A},
			},
		}
	}
	ch := chart.PieChart{
		Title:      title,
		TitleStyle: chart.Style{FontSize: 14},
		Background: chart.Style{Padding: chart.Box{Top: 40, Left: 20, Right: 20, Bottom: 20}},
		Height:     DefaultHeight,
		Width:      DefaultWidth,
		Values:     values,
	}
	var buf bytes.Buffer
	if err := ch.Render(chart.PNG, &buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func sumValues(data []BarDatum) float64 {
	var sum float64
	for _, d := range data {
		sum += d.Value
	}
	return sum
}

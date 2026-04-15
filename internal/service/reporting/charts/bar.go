package charts

import (
	"bytes"
	"fmt"

	"github.com/wcharczuk/go-chart/v2"
	"github.com/wcharczuk/go-chart/v2/drawing"
)

// BarDatum is a labelled value for a bar chart.
type BarDatum struct {
	Label string
	Value float64
}

// HorizontalBar renders a horizontal bar chart to PNG.
func HorizontalBar(title string, data []BarDatum) ([]byte, error) {
	if len(data) == 0 {
		return placeholder(title, "Sin datos")
	}
	bars := make([]chart.Value, len(data))
	for i, d := range data {
		c := ColorAt(i)
		bars[i] = chart.Value{
			Label: d.Label,
			Value: d.Value,
			Style: chart.Style{
				FillColor:   drawing.Color{R: c.R, G: c.G, B: c.B, A: c.A},
				StrokeColor: drawing.Color{R: c.R, G: c.G, B: c.B, A: c.A},
				StrokeWidth: 0,
			},
		}
	}
	ch := chart.BarChart{
		Title: title,
		TitleStyle: chart.Style{
			FontSize: 14,
		},
		Background: chart.Style{Padding: chart.Box{Top: 40, Left: 20, Right: 20, Bottom: 20}},
		Height:     DefaultHeight,
		Width:      DefaultWidth,
		BarWidth:   40,
		Bars:       bars,
		XAxis:      chart.Style{FontSize: 10},
		YAxis: chart.YAxis{
			Style: chart.Style{FontSize: 10},
			ValueFormatter: func(v interface{}) string {
				if f, ok := v.(float64); ok {
					return formatShort(f)
				}
				return fmt.Sprintf("%v", v)
			},
		},
	}
	var buf bytes.Buffer
	if err := ch.Render(chart.PNG, &buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// VerticalBar renders a vertical bar chart (same layout as HorizontalBar in go-chart).
func VerticalBar(title string, data []BarDatum) ([]byte, error) {
	return HorizontalBar(title, data)
}

// StackedBar renders multiple series stacked by category.
type StackedSeries struct {
	Name string
	Data []BarDatum // must share the same Labels in order
}

func StackedBar(title string, series []StackedSeries) ([]byte, error) {
	if len(series) == 0 || len(series[0].Data) == 0 {
		return placeholder(title, "Sin datos")
	}
	// Use a StackedBarChart
	labels := make([]string, len(series[0].Data))
	for i, d := range series[0].Data {
		labels[i] = d.Label
	}
	bars := make([]chart.StackedBar, len(labels))
	for i, lab := range labels {
		values := make([]chart.Value, len(series))
		for j, s := range series {
			var v float64
			if i < len(s.Data) {
				v = s.Data[i].Value
			}
			c := ColorAt(j)
			values[j] = chart.Value{
				Label: s.Name,
				Value: v,
				Style: chart.Style{
					FillColor:   drawing.Color{R: c.R, G: c.G, B: c.B, A: c.A},
					StrokeWidth: 0,
				},
			}
		}
		bars[i] = chart.StackedBar{Name: lab, Values: values}
	}
	ch := chart.StackedBarChart{
		Title:      title,
		TitleStyle: chart.Style{FontSize: 14},
		Background: chart.Style{Padding: chart.Box{Top: 40, Left: 20, Right: 20, Bottom: 20}},
		Height:     DefaultHeight,
		Width:      DefaultWidth,
		BarSpacing: 20,
		Bars:       bars,
	}
	var buf bytes.Buffer
	if err := ch.Render(chart.PNG, &buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func formatShort(v float64) string {
	abs := v
	if abs < 0 {
		abs = -abs
	}
	switch {
	case abs >= 1_000_000:
		return fmt.Sprintf("%.1fM", v/1_000_000)
	case abs >= 1_000:
		return fmt.Sprintf("%.1fk", v/1_000)
	default:
		return fmt.Sprintf("%.0f", v)
	}
}

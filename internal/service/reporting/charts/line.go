package charts

import (
	"bytes"
	"fmt"
	"time"

	"github.com/wcharczuk/go-chart/v2"
	"github.com/wcharczuk/go-chart/v2/drawing"
)

// TimeSeries holds a labelled time series.
type TimeSeries struct {
	Name   string
	Points []TimePoint
}

type TimePoint struct {
	Time  time.Time
	Value float64
}

// Lines renders one or more time series as a line chart to PNG.
func Lines(title string, series []TimeSeries) ([]byte, error) {
	total := 0
	for _, s := range series {
		total += len(s.Points)
	}
	if total == 0 {
		return placeholder(title, "Sin datos")
	}

	chSeries := make([]chart.Series, 0, len(series))
	for i, s := range series {
		if len(s.Points) == 0 {
			continue
		}
		xv := make([]time.Time, len(s.Points))
		yv := make([]float64, len(s.Points))
		for j, p := range s.Points {
			xv[j] = p.Time
			yv[j] = p.Value
		}
		c := ColorAt(i)
		chSeries = append(chSeries, chart.TimeSeries{
			Name:    s.Name,
			XValues: xv,
			YValues: yv,
			Style: chart.Style{
				StrokeColor: drawing.Color{R: c.R, G: c.G, B: c.B, A: c.A},
				StrokeWidth: 2.0,
			},
		})
	}

	ch := chart.Chart{
		Title:      title,
		TitleStyle: chart.Style{FontSize: 14},
		Background: chart.Style{Padding: chart.Box{Top: 40, Left: 20, Right: 20, Bottom: 20}},
		Height:     DefaultHeight,
		Width:      DefaultWidth,
		XAxis: chart.XAxis{
			Style: chart.Style{FontSize: 10},
			ValueFormatter: chart.TimeValueFormatterWithFormat("02/01"),
		},
		YAxis: chart.YAxis{
			Style: chart.Style{FontSize: 10},
			ValueFormatter: func(v interface{}) string {
				if f, ok := v.(float64); ok {
					return formatShort(f)
				}
				return fmt.Sprintf("%v", v)
			},
		},
		Series: chSeries,
	}
	if len(chSeries) > 1 {
		ch.Elements = []chart.Renderable{chart.LegendThin(&ch)}
	}
	var buf bytes.Buffer
	if err := ch.Render(chart.PNG, &buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

package charts

import (
	"bytes"

	"github.com/wcharczuk/go-chart/v2"
)

// placeholder renders a simple chart with just a title when data is empty.
// It avoids callers having to special-case the empty state.
func placeholder(title, message string) ([]byte, error) {
	ch := chart.Chart{
		Title:      title + " — " + message,
		TitleStyle: chart.Style{FontSize: 14},
		Background: chart.Style{Padding: chart.Box{Top: 40, Left: 20, Right: 20, Bottom: 20}},
		Height:     DefaultHeight,
		Width:      DefaultWidth,
		Series: []chart.Series{
			chart.ContinuousSeries{
				XValues: []float64{0, 1},
				YValues: []float64{0, 0},
			},
		},
	}
	var buf bytes.Buffer
	if err := ch.Render(chart.PNG, &buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

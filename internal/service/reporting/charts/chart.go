// Package charts renders PNG charts used to embed in PDF reports.
package charts

import "image/color"

// Palette used across charts — aligned with the brand colors of Ricitelli.
var Palette = []color.RGBA{
	{R: 78, G: 7, B: 7, A: 255},     // wine red
	{R: 184, G: 134, B: 11, A: 255}, // gold
	{R: 85, G: 107, B: 47, A: 255},  // olive
	{R: 168, G: 85, B: 247, A: 255}, // purple
	{R: 59, G: 130, B: 246, A: 255}, // blue
	{R: 239, G: 68, B: 68, A: 255},  // red
	{R: 34, G: 197, B: 94, A: 255},  // green
	{R: 249, G: 115, B: 22, A: 255}, // orange
	{R: 139, G: 115, B: 85, A: 255}, // taupe
	{R: 20, G: 184, B: 166, A: 255}, // teal
}

// ColorAt returns a palette color cycling through indices.
func ColorAt(i int) color.RGBA {
	if len(Palette) == 0 {
		return color.RGBA{0, 0, 0, 255}
	}
	return Palette[i%len(Palette)]
}

// DefaultWidth / DefaultHeight are the output pixel dimensions of generated PNGs.
const (
	DefaultWidth  = 900
	DefaultHeight = 500
)

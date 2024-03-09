package interpreter

// From https://rosepinetheme.com/palette/ingredients/

import (
	"image/color"
)

type Color struct {
	R, G, B uint8
}

var Colors = map[string]Color{
	"Base":    {25, 23, 36},
	"Surface": {31, 29, 46},
	"Overlay": {38, 35, 58},
	"Muted":   {110, 106, 134},
	"Subtle":  {144, 140, 170},
	"Text":    {224, 222, 244},
	"Love":    {235, 111, 146},
	"Gold":    {246, 193, 119},
	"Rose":    {235, 188, 186},
	"Pine":    {49, 116, 143},
	"Foam":    {156, 207, 216},
	"Iris":    {196, 167, 231},
}

func (c Color) RGBA() color.RGBA {
	return color.RGBA{c.R, c.G, c.B, 255}
}

package chip8

// From https://rosepinetheme.com/palette/ingredients/

import (
	"image/color"
)

type Color struct {
	R, G, B uint8
}

var (
	Base    = Color{25, 23, 36}
	Surface = Color{31, 29, 46}
	Overlay = Color{38, 35, 58}
	Muted   = Color{110, 106, 134}
	Subtle  = Color{144, 140, 170}
	Text    = Color{224, 222, 244}
	Love    = Color{235, 111, 146}
	Gold    = Color{246, 193, 119}
	Rose    = Color{235, 188, 186}
	Pine    = Color{49, 116, 143}
	Foam    = Color{156, 207, 216}
	Iris    = Color{196, 167, 231}
)

func (c Color) RGBA() color.RGBA {
	return color.RGBA{c.R, c.G, c.B, 255}
}

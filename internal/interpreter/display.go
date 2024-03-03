package interpreter

import "image/color"

type Display struct {
	content  [DisplayWidth][DisplayHeight]byte
	offColor color.Color
	onColor  color.Color
}

func (d *Display) clear() {
	for x := 0; x < DisplayWidth; x++ {
		for y := 0; y < DisplayHeight; y++ {
			d.content[x][y] = 0
		}
	}
}

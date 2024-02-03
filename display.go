package main

import "image/color"

type Display struct {
	content  [displayWidth][displayHeight]byte
	offColor color.Color
	onColor  color.Color
}

func (d *Display) clear() {
	for x := 0; x < displayWidth; x++ {
		for y := 0; y < displayHeight; y++ {
			d.content[x][y] = 0
		}
	}
}

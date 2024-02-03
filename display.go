package main

import "image/color"

type Display struct {
	content  [displayWidth][displayHeight]byte
	offColor color.Color
	onColor  color.Color
}

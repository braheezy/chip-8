package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

const (
	displayWidth  = 64
	displayHeight = 32
	// So it can be seen on modern displays
	displayScaleFactor = 10
	// Set TPS of execution loop. Controls how many times per second Update() is run.
	cpuSpeed = 60
)

func (ch8 *CHIP8) Update() error {
	if ebiten.IsKeyPressed(ebiten.KeyEscape) {
		return ebiten.Termination
	}
	// Execute instructions
	for int(ch8.pc) < ch8.programSize {
		// println("hi")
		firstNib, secondNib := ch8.readNextInstruction()
		switch firstNib {
		case 0x00:
			switch secondNib {
			case 0xE0:
				// Clear display
				ch8.display.clear()
			}
		default:
			fmt.Printf("Unknown instruction starting with: %X\n", firstNib)
		}
	}
	if int(ch8.pc) == ch8.programSize {
		return ebiten.Termination
	}
	return nil
}
func (chip8 *CHIP8) Draw(screen *ebiten.Image) {
	// Effectively clear the screen for new draw.
	screen.Fill(chip8.display.offColor)

	// Iterate over CHIP-8 display data.
	for x := 0; x < displayWidth; x++ {
		for y := 0; y < displayHeight; y++ {
			if chip8.display.content[x][y] != 0 {
				// Draw a filled rectangle for each set CHIP-8 pixel
				vector.DrawFilledRect(
					screen,
					float32(x*displayScaleFactor),
					float32(y*displayScaleFactor),
					displayScaleFactor,
					displayScaleFactor,
					chip8.display.onColor,
					false,
				)
			}
		}
	}
}

func (chip8 *CHIP8) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return displayWidth * displayScaleFactor, displayHeight * displayScaleFactor
}

func main() {
	// Read positional argument. Must be provided and must be a file.
	// If no argument is provided, print usage and exit.
	if len(os.Args) < 2 {
		log.Fatal("No file provided.")
	}

	chipFilePath := os.Args[1]
	chipFileName := filepath.Base(chipFilePath)
	chipData, err := os.ReadFile(chipFilePath)
	if err != nil {
		log.Fatal(err)
	}

	chip8 := NewCHIP8()
	chip8.programSize = len(chipData) + programStartAddress

	// Load program into memory.
	copy(chip8.memory[programStartAddress:], chipData)

	ebiten.SetWindowSize(displayWidth*displayScaleFactor, displayHeight*displayScaleFactor)
	ebiten.SetWindowTitle(chipFileName)
	ebiten.SetTPS(cpuSpeed)
	if err := ebiten.RunGame(chip8); err != nil && err != ebiten.Termination {
		log.Fatal(err)
	}

}

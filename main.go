package main

import (
	"flag"
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
	// Limit how many cycles the program is run for. For debug purposes.
	cycleLimit = -1
)

var (
	currentCycle = 0
)

func (ch8 *CHIP8) Update() error {
	if ebiten.IsKeyPressed(ebiten.KeyEscape) {
		return ebiten.Termination
	}
	if currentCycle == cycleLimit {
		return nil
	}
	ch8.stepInterpreter()
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

var debugFlag bool

func init() {
	flag.BoolVar(&debugFlag, "debug", false, "Show debug messages")
}

func main() {
	flag.Parse()
	if len(flag.Args()) < 1 {
		log.Fatal("No file provided.")
	}

	chipFilePath := flag.Arg(0)
	chipFileName := filepath.Base(chipFilePath)
	chipData, err := os.ReadFile(chipFilePath)
	if err != nil {
		log.Fatal(err)
	}

	if debugFlag {
		logLevel = Debug
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

package chip8

import (
	"time"

	"github.com/charmbracelet/log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

type App struct {
	Chip8   *CHIP8
	Logger  *log.Logger
	Options CHIP8Options
}

// NewDefaultApp creates a new App with default options.
func NewDefaultApp(romData *[]byte) *App {
	return &App{
		Chip8:  NewCHIP8WithOptions(romData, DefaultOptions()),
		Logger: log.Default(),
	}
}

// NewAppWithOptions creates a new App with specified options.
func NewAppWithOptions(romData *[]byte, options CHIP8Options) *App {
	return &App{
		Chip8:  NewCHIP8WithOptions(romData, options),
		Logger: log.Default(),
	}
}

func (app *App) Update() error {
	if ebiten.IsKeyPressed(ebiten.KeyEscape) {
		return ebiten.Termination
	}
	if currentCycle == app.Chip8.Options.CycleLimit {
		return nil
	}

	// Calculate elapsed time since last timer update
	elapsed := time.Since(lastDelayTimerUpdate)
	if app.Chip8.delayTimer > 0 {
		if elapsed >= decrementInterval {
			app.Chip8.delayTimer--
			elapsed -= decrementInterval
			lastDelayTimerUpdate = lastDelayTimerUpdate.Add(decrementInterval)
		}
	}

	elapsed = time.Since(lastSoundTimerUpdate)
	if app.Chip8.soundTimer > 0 {
		app.Chip8.beep.Play()
		if elapsed >= decrementInterval {
			app.Chip8.soundTimer--
			elapsed -= decrementInterval
			lastSoundTimerUpdate = lastSoundTimerUpdate.Add(decrementInterval)
		}
	} else {
		if app.Chip8.beep.IsPlaying() {
			app.Chip8.beep.Pause()
			app.Chip8.beep.SetPosition(0)
		}
	}

	// Handle input
	var keys []ebiten.Key
	keys = inpututil.AppendPressedKeys(keys)
	if len(keys) > 0 {
		// For any pressed keys, convert them to hex
		var keypresses []byte
		for _, key := range keys {
			keypress, err := keyToHex(key)
			if err == nil {
				keypresses = append(keypresses, keypress)
			}
		}
		app.Chip8.pressedKeys = keypresses
		app.Chip8.dirtyKeys = true
	} else {
		if !app.Chip8.dirtyKeys {
			app.Chip8.pressedKeys = []byte{}
		}
	}

	app.Chip8.stepInterpreter()
	if int(app.Chip8.pc) == app.Chip8.programSize {
		return ebiten.Termination
	}
	return nil
}

func (app *App) Draw(screen *ebiten.Image) {
	// Iterate over CHIP-8 display data.
	for x := 0; x < DisplayWidth; x++ {
		for y := 0; y < DisplayHeight; y++ {
			if app.Chip8.display.content[x][y] != 0 {
				// Draw a filled rectangle for each set CHIP-8 pixel
				vector.DrawFilledRect(
					screen,
					float32(x*app.Chip8.Options.DisplayScaleFactor),
					float32(y*app.Chip8.Options.DisplayScaleFactor),
					float32(app.Chip8.Options.DisplayScaleFactor),
					float32(app.Chip8.Options.DisplayScaleFactor),
					app.Chip8.display.onColor,
					false,
				)
			} else {
				// Draw a rectangle for each unset CHIP-8 pixel
				vector.DrawFilledRect(
					screen,
					float32(x*app.Chip8.Options.DisplayScaleFactor),
					float32(y*app.Chip8.Options.DisplayScaleFactor),
					float32(app.Chip8.Options.DisplayScaleFactor),
					float32(app.Chip8.Options.DisplayScaleFactor),
					app.Chip8.display.offColor,
					false,
				)
			}
		}
	}
}

func (app *App) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return DisplayWidth * app.Chip8.Options.DisplayScaleFactor, DisplayHeight * app.Chip8.Options.DisplayScaleFactor
}

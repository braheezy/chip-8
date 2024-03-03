package interpreter

import (
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

func (app *CHIP8) Update() error {
	if ebiten.IsKeyPressed(ebiten.KeyEscape) {
		return ebiten.Termination
	}
	// Calculate elapsed time since last timer update
	elapsed := time.Since(lastDelayTimerUpdate)
	if app.delayTimer > 0 {
		if elapsed >= decrementInterval {
			app.delayTimer--
			elapsed -= decrementInterval
			lastDelayTimerUpdate = lastDelayTimerUpdate.Add(decrementInterval)
		}
	}

	elapsed = time.Since(lastSoundTimerUpdate)
	if app.soundTimer > 0 {
		app.beep.Play()
		if elapsed >= decrementInterval {
			app.soundTimer--
			elapsed -= decrementInterval
			lastSoundTimerUpdate = lastSoundTimerUpdate.Add(decrementInterval)
		}
	} else {
		if app.beep.IsPlaying() {
			app.beep.Pause()
			app.beep.SetPosition(0)
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
		app.pressedKeys = keypresses
		app.dirtyKeys = true
	} else {
		if !app.dirtyKeys {
			app.pressedKeys = []byte{}
		}
	}

	app.stepInterpreter()
	if int(app.pc) == app.programSize {
		return ebiten.Termination
	}
	return nil
}

func (app *CHIP8) Draw(screen *ebiten.Image) {
	// Iterate over CHIP-8 display data.
	for x := 0; x < DisplayWidth; x++ {
		for y := 0; y < DisplayHeight; y++ {
			if app.display.content[x][y] != 0 {
				// Draw a filled rectangle for each set CHIP-8 pixel
				vector.DrawFilledRect(
					screen,
					float32(x*app.Options.DisplayScaleFactor),
					float32(y*app.Options.DisplayScaleFactor),
					float32(app.Options.DisplayScaleFactor),
					float32(app.Options.DisplayScaleFactor),
					app.display.onColor,
					false,
				)
			} else {
				// Draw a rectangle for each unset CHIP-8 pixel
				vector.DrawFilledRect(
					screen,
					float32(x*app.Options.DisplayScaleFactor),
					float32(y*app.Options.DisplayScaleFactor),
					float32(app.Options.DisplayScaleFactor),
					float32(app.Options.DisplayScaleFactor),
					app.display.offColor,
					false,
				)
			}
		}
	}
}

func (app *CHIP8) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return DisplayWidth * app.Options.DisplayScaleFactor, DisplayHeight * app.Options.DisplayScaleFactor
}

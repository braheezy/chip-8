package interpreter

import (
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

func (chip8 *CHIP8) Update() error {
	if ebiten.IsKeyPressed(ebiten.KeyEscape) {
		return ebiten.Termination
	}
	// Calculate elapsed time since last timer update
	elapsed := time.Since(lastDelayTimerUpdate)
	if chip8.delayTimer > 0 {
		if elapsed >= decrementInterval {
			chip8.delayTimer--
			elapsed -= decrementInterval
			lastDelayTimerUpdate = lastDelayTimerUpdate.Add(decrementInterval)
		}
	}

	elapsed = time.Since(lastSoundTimerUpdate)
	if chip8.soundTimer > 0 {
		chip8.beep.Play()
		if elapsed >= decrementInterval {
			chip8.soundTimer--
			elapsed -= decrementInterval
			lastSoundTimerUpdate = lastSoundTimerUpdate.Add(decrementInterval)
		}
	} else {
		if chip8.beep.IsPlaying() {
			chip8.beep.Pause()
			chip8.beep.SetPosition(0)
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
		chip8.pressedKeys = keypresses
		chip8.dirtyKeys = true
	} else {
		if !chip8.dirtyKeys {
			chip8.pressedKeys = []byte{}
		}
	}

	chip8.stepInterpreter()
	if int(chip8.pc) == chip8.programSize {
		return ebiten.Termination
	}
	return nil
}

func (chip8 *CHIP8) Draw(screen *ebiten.Image) {
	// Iterate over CHIP-8 display data.
	for x := 0; x < DisplayWidth; x++ {
		for y := 0; y < DisplayHeight; y++ {
			if chip8.display.content[x][y] != 0 {
				// Draw a filled rectangle for each set CHIP-8 pixel
				vector.DrawFilledRect(
					screen,
					float32(x*chip8.Options.DisplayScaleFactor),
					float32(y*chip8.Options.DisplayScaleFactor),
					float32(chip8.Options.DisplayScaleFactor),
					float32(chip8.Options.DisplayScaleFactor),
					chip8.display.onColor,
					false,
				)
			} else {
				// Draw a rectangle for each unset CHIP-8 pixel
				vector.DrawFilledRect(
					screen,
					float32(x*chip8.Options.DisplayScaleFactor),
					float32(y*chip8.Options.DisplayScaleFactor),
					float32(chip8.Options.DisplayScaleFactor),
					float32(chip8.Options.DisplayScaleFactor),
					chip8.display.offColor,
					false,
				)
			}
		}
	}
}

func (chip8 *CHIP8) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return DisplayWidth * chip8.Options.DisplayScaleFactor, DisplayHeight * chip8.Options.DisplayScaleFactor
}

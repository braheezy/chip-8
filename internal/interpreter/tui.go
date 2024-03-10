package interpreter

import (
	"errors"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func RunTUI(chip8 *CHIP8, filename string) {
	chip8.Logger.Info("Running TUI", "romFile", filename)

	app := &App{Chip8: chip8}

	p := tea.NewProgram(app)
	p.SetWindowTitle(filename)
	if _, err := p.Run(); err != nil {
		chip8.Logger.Fatalf("Could not start program :(\n%v\n", err)
	}

}

// Hack in a delay because BubbleTea doesn't do real time input(?)
// If set to 0, all inputs are "held" continuously
// This value seems to hit the sweet spot.
const defaultInputDelay = 85

type App struct {
	Chip8             *CHIP8
	CurrentInputDelay int
	terminalHeight    int
	terminalWidth     int
}

type execMsg interface{}

func (app *App) Init() tea.Cmd {
	return func() tea.Msg {
		return execMsg(true)
	}
}

func (app *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		needsRepaint := false
		if msg.Width < app.terminalWidth {
			needsRepaint = true
		}
		app.terminalHeight = msg.Height
		app.terminalWidth = msg.Width

		if needsRepaint {
			return app, tea.ClearScreen
		}

	// User pressed a key
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" || msg.String() == "esc" {
			return app, tea.Quit
		} else {
			keypress, err := teaKeyToHex(msg)
			if err == nil {
				app.Chip8.Logger.Warnf("user pressing %X", keypress)
				app.Chip8.pressedKeys = []byte{keypress}
				app.Chip8.dirtyKeys = true
				app.CurrentInputDelay = defaultInputDelay
			}
		}
	}
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

	// CurrentInputDelay prevents the pressed key from being cleared too quickly
	if !app.Chip8.dirtyKeys && app.CurrentInputDelay == 0 {
		app.Chip8.pressedKeys = []byte{}
	} else {
		app.CurrentInputDelay--
	}

	app.Chip8.stepInterpreter()

	if int(app.Chip8.pc) == app.Chip8.programSize {
		return app, tea.Quit
	}

	return app, func() tea.Msg {
		return execMsg(true)
	}
}

func (app *App) View() string {
	view := strings.Builder{}
	for y := 0; y < DisplayHeight; y++ {
		for x := 0; x < DisplayWidth; x++ {
			if app.Chip8.display.content[x][y] != 0 {
				s := lipgloss.NewStyle().SetString("  ").Background(Colors[app.Chip8.Options.OnColor])
				view.WriteString(s.String())
			} else {
				s := lipgloss.NewStyle().SetString("  ").Background(Colors[app.Chip8.Options.OffColor])
				view.WriteString(s.String())
			}
		}
		view.WriteRune('\n')
	}
	return view.String()
}

// Convert keypad key to hex value
func teaKeyToHex(key tea.KeyMsg) (byte, error) {
	var hexValue byte
	switch key.String() {
	case "x":
		hexValue = 0x0
	case "1":
		hexValue = 0x1
	case "2":
		hexValue = 0x2
	case "3":
		hexValue = 0x3
	case "4":
		hexValue = 0xC
	case "q":
		hexValue = 0x4
	case "w":
		hexValue = 0x5
	case "e":
		hexValue = 0x6
	case "r":
		hexValue = 0xD
	case "a":
		hexValue = 0x7
	case "s":
		hexValue = 0x8
	case "d":
		hexValue = 0x9
	case "f":
		hexValue = 0xE
	case "z":
		hexValue = 0xA
	case "c":
		hexValue = 0xB
	case "v":
		hexValue = 0xF
	default:
		return 0, errors.New("unsupported key")
	}
	return hexValue, nil
}

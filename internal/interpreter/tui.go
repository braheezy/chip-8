package interpreter

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func RunTUI(chip8 *CHIP8, filename string) {
	chip8.Logger.Info("Running TUI", "romFile", filename)

	app := &App{chip8}

	p := tea.NewProgram(app)
	p.SetWindowTitle(filename)
	if _, err := p.Run(); err != nil {
		chip8.Logger.Fatalf("Could not start program :(\n%v\n", err)
	}

}

type App struct {
	Chip8 *CHIP8
}

type execMsg bool

func (app *App) Init() tea.Cmd {
	return func() tea.Msg {
		return execMsg(true)
	}
}

func (app *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			return app, tea.Quit
		}
	case execMsg:
		app.Chip8.stepInterpreter()
		return app, func() tea.Msg {
			return execMsg(true)
		}
	}
	return app, nil
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

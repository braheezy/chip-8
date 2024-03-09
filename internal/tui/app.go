package tui

import "github.com/charmbracelet/log"

func Run(romFilePath string, logger *log.Logger) {
	logger.Info("Running TUI", "romFile", romFilePath)
}

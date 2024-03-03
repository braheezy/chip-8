package cmd

import (
	"os"

	"github.com/charmbracelet/log"
)

var logger *log.Logger

func setupLogger() {
	logger = log.New(os.Stdout)

	if debug {
		logger.SetLevel(log.DebugLevel)
	}
}

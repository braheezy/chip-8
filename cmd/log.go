package cmd

import (
	"os"

	"github.com/charmbracelet/log"
)

func newDefaultLogger() *log.Logger {
	return initLogger(log.InfoLevel)
}

func initLogger(level log.Level) *log.Logger {
	logger := log.New(os.Stdout)
	logger.SetLevel(level)

	return logger
}

package chip8

import (
	"log"

	clog "github.com/charmbracelet/log"
)

type LogLevel int

const (
	Debug LogLevel = iota
	Info
	Warn
	Error
)

var logLevel = Info

var logger *clog.Logger

func debug(format string, args ...interface{}) {
	if logger == nil {
		if logLevel == Debug {
			log.Printf(format, args...)
		}
	} else {
		logger.Debug(format, args...)
	}
}

func warn(format string, args ...interface{}) {
	if logger == nil {
		if logLevel <= Warn {
			log.Printf(format, args...)
		}
	} else {
		logger.Warn(format, args...)
	}
}

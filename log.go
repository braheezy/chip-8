package main

import "log"

type LogLevel int

const (
	Debug LogLevel = iota
	Info
	Warn
	Error
)

var logLevel = Info

func debug(format string, args ...interface{}) {
	if logLevel == Debug {
		log.Printf(format, args...)
	}
}

func warn(format string, args ...interface{}) {
	if logLevel <= Warn {
		log.Printf(format, args...)
	}
}

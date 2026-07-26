package main

import "log"

var enableServerLogs bool

func serverLogf(format string, args ...any) {
	if enableServerLogs {
		log.Printf(format, args...)
	}
}

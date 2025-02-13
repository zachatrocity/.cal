package logger

import (
	"log"
	"os"
)

var debug = false

func init() {
	debug = os.Getenv("DEV_MODE") == "true"
}

// Debug logs a debug message if DEV_MODE is enabled
func Debug(format string, v ...interface{}) {
	if debug {
		log.Printf("[DEBUG] "+format, v...)
	}
}

// Error logs an error message
func Error(format string, v ...interface{}) {
	log.Printf("[ERROR] "+format, v...)
}

// Info logs an info message
func Info(format string, v ...interface{}) {
	log.Printf("[INFO] "+format, v...)
}

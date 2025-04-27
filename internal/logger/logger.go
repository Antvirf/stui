package logger

import (
	"fmt"
	"log"
	"os"

	"github.com/antvirf/stui/internal/config"
)

var (
	std = log.New(os.Stderr, "", log.LstdFlags)
)

// Printf calls Output to print to the standard logger if quiet is false.
func Printf(format string, v ...interface{}) {
	if config.LogLevel >= config.LOG_LEVEL_INFO {
		std.Printf(format, v...)
	}
}

// Println calls Output to print to the standard logger if quiet is false.
func Println(v ...interface{}) {
	if config.LogLevel >= config.LOG_LEVEL_INFO {
		std.Println(v...)
	}
}

// Debug prints debug information if quiet is false.
// This provides an additional convenience method for debug logs.
func Debug(v ...interface{}) {
	if config.LogLevel >= config.LOG_LEVEL_DEBUG {
		std.Print(fmt.Sprintf("[DEBUG] %v", fmt.Sprint(v...)))
	}
}

// Debugf prints formatted debug information if quiet is false.
func Debugf(format string, v ...interface{}) {
	if config.LogLevel >= config.LOG_LEVEL_DEBUG {
		std.Print(fmt.Sprintf("[DEBUG] %v", fmt.Sprintf(format, v...)))
	}
}

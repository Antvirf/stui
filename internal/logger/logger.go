package logger

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"sync"

	"github.com/antvirf/stui/internal/config"
)

var (
	// Standard logger for direct output
	stdLogger *log.Logger

	// Buffered logger
	bufferLogger *log.Logger

	// Buffer to store logs
	logBuffer *bytes.Buffer

	// Mutex to protect buffer access
	bufferMutex sync.Mutex

	// Flag to control buffering
	isBuffered bool
)

func init() {
	logBuffer = &bytes.Buffer{}
	stdLogger = log.New(os.Stderr, "", log.LstdFlags)
	bufferLogger = log.New(logBuffer, "", log.LstdFlags)
}

// EnableBuffering turns on log buffering
func EnableBuffering() {
	bufferMutex.Lock()
	defer bufferMutex.Unlock()
	isBuffered = true
}

// LogFlush prints all buffered logs to stderr and clears the buffer
func LogFlush() {
	bufferMutex.Lock()
	defer bufferMutex.Unlock()

	if logBuffer.Len() > 0 {
		io.Copy(os.Stderr, logBuffer)
		logBuffer.Reset()
	}
}

// getLogger returns the appropriate logger based on buffering state
func getLogger() *log.Logger {
	if isBuffered {
		return bufferLogger
	}
	return stdLogger
}

// Printf calls Output to print to the appropriate logger
func Printf(format string, v ...interface{}) {
	if config.LogLevel >= config.LOG_LEVEL_INFO {
		bufferMutex.Lock()
		defer bufferMutex.Unlock()
		getLogger().Printf(format, v...)
	}
}

// Println calls Output to print to the appropriate logger
func Println(v ...interface{}) {
	if config.LogLevel >= config.LOG_LEVEL_INFO {
		bufferMutex.Lock()
		defer bufferMutex.Unlock()
		getLogger().Println(v...)
	}
}

// Debug prints debug information
func Debug(v ...interface{}) {
	if config.LogLevel >= config.LOG_LEVEL_DEBUG {
		bufferMutex.Lock()
		defer bufferMutex.Unlock()
		getLogger().Printf("[DEBUG] %v", fmt.Sprint(v...))
	}
}

// Debugf prints formatted debug information
func Debugf(format string, v ...interface{}) {
	if config.LogLevel >= config.LOG_LEVEL_DEBUG {
		bufferMutex.Lock()
		defer bufferMutex.Unlock()
		getLogger().Printf("[DEBUG] %v", fmt.Sprintf(format, v...))
	}
}

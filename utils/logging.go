package utils

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
)

// LogLevel represents different logging levels
type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	ERROR
)

var (
	// Current log level
	currentLevel = INFO
	
	// Logger instances for different levels
	debugLogger = log.New(os.Stdout, "DEBUG: ", log.Ldate|log.Ltime|log.Lmicroseconds)
	infoLogger  = log.New(os.Stdout, "INFO:  ", log.Ldate|log.Ltime|log.Lmicroseconds)
	errorLogger = log.New(os.Stderr, "ERROR: ", log.Ldate|log.Ltime|log.Lmicroseconds)
)

// SetLogLevel sets the current logging level
func SetLogLevel(level LogLevel) {
	currentLevel = level
}

// getFileAndLine returns the source file and line number
func getFileAndLine() string {
	_, file, line, ok := runtime.Caller(2)
	if !ok {
		return "???"
	}
	// Get just the file name without the full path
	file = filepath.Base(file)
	return fmt.Sprintf("%s:%d", file, line)
}

// formatMessage formats a message with its source location
func formatMessage(msg string, args ...interface{}) string {
	location := getFileAndLine()
	formattedMsg := fmt.Sprintf(msg, args...)
	return fmt.Sprintf("[%s] %s", location, formattedMsg)
}

// Debug logs a debug message if debug level is enabled
func Debug(format string, args ...interface{}) {
	if currentLevel <= DEBUG {
		debugLogger.Output(2, formatMessage(format, args...))
	}
}

// Info logs an info message
func Info(format string, args ...interface{}) {
	if currentLevel <= INFO {
		infoLogger.Output(2, formatMessage(format, args...))
	}
}

// Error logs an error message
func Error(format string, args ...interface{}) {
	if currentLevel <= ERROR {
		errorLogger.Output(2, formatMessage(format, args...))
	}
}

// EnableDebugLogging enables debug level logging
func EnableDebugLogging() {
	currentLevel = DEBUG
}

// DisableDebugLogging sets logging level to INFO
func DisableDebugLogging() {
	currentLevel = INFO
}

// IsDebugEnabled returns whether debug logging is enabled
func IsDebugEnabled() bool {
	return currentLevel == DEBUG
}

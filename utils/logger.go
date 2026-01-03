package utils

import (
	"fmt"
	"log"
	"os"
	"time"
)

// LogLevel definiert die Log-Stufen
type LogLevel int

const (
	LogLevelDebug LogLevel = iota
	LogLevelInfo
	LogLevelWarn
	LogLevelError
	LogLevelFatal
)

// Logger ist ein einfacher Logger mit Modulnamen
type Logger struct {
	name  string
	level LogLevel
}

var globalLogLevel = LogLevelInfo

// SetGlobalLogLevel setzt das globale Log-Level
func SetGlobalLogLevel(level LogLevel) {
	globalLogLevel = level
}

// SetGlobalLogLevelFromString setzt das Log-Level aus einem String
func SetGlobalLogLevelFromString(level string) {
	switch level {
	case "DEBUG":
		globalLogLevel = LogLevelDebug
	case "INFO":
		globalLogLevel = LogLevelInfo
	case "WARN":
		globalLogLevel = LogLevelWarn
	case "ERROR":
		globalLogLevel = LogLevelError
	case "FATAL":
		globalLogLevel = LogLevelFatal
	default:
		globalLogLevel = LogLevelInfo
	}
}

// NewLogger erstellt einen neuen Logger
func NewLogger(name string) *Logger {
	return &Logger{
		name:  name,
		level: globalLogLevel,
	}
}

// Debug loggt eine Debug-Nachricht
func (l *Logger) Debug(format string, v ...interface{}) {
	if l.shouldLog(LogLevelDebug) {
		l.log("DEBUG", format, v...)
	}
}

// Info loggt eine Info-Nachricht
func (l *Logger) Info(format string, v ...interface{}) {
	if l.shouldLog(LogLevelInfo) {
		l.log("INFO", format, v...)
	}
}

// Warn loggt eine Warnung
func (l *Logger) Warn(format string, v ...interface{}) {
	if l.shouldLog(LogLevelWarn) {
		l.log("WARN", format, v...)
	}
}

// Error loggt einen Fehler
func (l *Logger) Error(format string, v ...interface{}) {
	if l.shouldLog(LogLevelError) {
		l.log("ERROR", format, v...)
	}
}

// Fatal loggt einen fatalen Fehler und beendet das Programm
func (l *Logger) Fatal(format string, v ...interface{}) {
	l.log("FATAL", format, v...)
	os.Exit(1)
}

// shouldLog prüft ob das Log-Level hoch genug ist
func (l *Logger) shouldLog(level LogLevel) bool {
	return level >= globalLogLevel
}

// log ist die interne Log-Funktion
func (l *Logger) log(level string, format string, v ...interface{}) {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	prefix := fmt.Sprintf("[%s] [%s] [%s]", timestamp, level, l.name)
	message := fmt.Sprintf(format, v...)
	log.Printf("%s: %s", prefix, message)
}

// LogToFile konfiguriert das Logging in eine Datei
func LogToFile(filepath string) error {
	file, err := os.OpenFile(filepath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return err
	}
	log.SetOutput(file)
	return nil
}

// LogToStdout setzt das Logging auf Stdout zurück
func LogToStdout() {
	log.SetOutput(os.Stdout)
}

// SimpleLog ist eine einfache Log-Funktion ohne Logger-Instanz
func SimpleLog(level, message string) {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	log.Printf("[%s] [%s]: %s", timestamp, level, message)
}

// DebugLog loggt eine Debug-Nachricht (globale Funktion)
func DebugLog(format string, v ...interface{}) {
	if globalLogLevel <= LogLevelDebug {
		SimpleLog("DEBUG", fmt.Sprintf(format, v...))
	}
}

// InfoLog loggt eine Info-Nachricht (globale Funktion)
func InfoLog(format string, v ...interface{}) {
	if globalLogLevel <= LogLevelInfo {
		SimpleLog("INFO", fmt.Sprintf(format, v...))
	}
}

// ErrorLog loggt einen Fehler (globale Funktion)
func ErrorLog(format string, v ...interface{}) {
	if globalLogLevel <= LogLevelError {
		SimpleLog("ERROR", fmt.Sprintf(format, v...))
	}
}

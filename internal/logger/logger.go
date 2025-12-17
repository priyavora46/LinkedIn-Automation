package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"
)

type Level int

const (
	DEBUG Level = iota
	INFO
	WARN
	ERROR
)

type Logger struct {
	level    Level
	infoLog  *log.Logger
	warnLog  *log.Logger
	errorLog *log.Logger
	debugLog *log.Logger
	file     *os.File
}

func New(level string, logFile string, console bool) (*Logger, error) {
	l := &Logger{}

	// Parse level
	switch level {
	case "debug":
		l.level = DEBUG
	case "info":
		l.level = INFO
	case "warn":
		l.level = WARN
	case "error":
		l.level = ERROR
	default:
		l.level = INFO
	}

	// Create writers
	writers := []io.Writer{}

	if console {
		writers = append(writers, os.Stdout)
	}

	if logFile != "" {
		// Create log directory
		dir := filepath.Dir(logFile)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create log directory: %w", err)
		}

		file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			return nil, fmt.Errorf("failed to open log file: %w", err)
		}
		l.file = file
		writers = append(writers, file)
	}

	writer := io.MultiWriter(writers...)

	// Create loggers
	l.debugLog = log.New(writer, "[DEBUG] ", log.Ldate|log.Ltime|log.Lshortfile)
	l.infoLog = log.New(writer, "[INFO]  ", log.Ldate|log.Ltime)
	l.warnLog = log.New(writer, "[WARN]  ", log.Ldate|log.Ltime)
	l.errorLog = log.New(writer, "[ERROR] ", log.Ldate|log.Ltime|log.Lshortfile)

	return l, nil
}

func (l *Logger) Debug(format string, v ...interface{}) {
	if l.level <= DEBUG {
		l.debugLog.Output(2, fmt.Sprintf(format, v...))
	}
}

func (l *Logger) Info(format string, v ...interface{}) {
	if l.level <= INFO {
		l.infoLog.Output(2, fmt.Sprintf(format, v...))
	}
}

func (l *Logger) Warn(format string, v ...interface{}) {
	if l.level <= WARN {
		l.warnLog.Output(2, fmt.Sprintf(format, v...))
	}
}

func (l *Logger) Error(format string, v ...interface{}) {
	if l.level <= ERROR {
		l.errorLog.Output(2, fmt.Sprintf(format, v...))
	}
}

func (l *Logger) Close() error {
	if l.file != nil {
		return l.file.Close()
	}
	return nil
}

func (l *Logger) LogAction(action string, details map[string]interface{}) {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	msg := fmt.Sprintf("[ACTION] %s - %s", timestamp, action)
	for k, v := range details {
		msg += fmt.Sprintf(" | %s=%v", k, v)
	}
	l.Info(msg)
}

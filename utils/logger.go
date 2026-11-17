package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type LogLevel string

const (
	DEBUG LogLevel = "DEBUG"
	INFO  LogLevel = "INFO"
	WARN  LogLevel = "WARN"
	ERROR LogLevel = "ERROR"
)

const (
	ColorReset  = "\033[0m"
	ColorRed    = "\033[31m"
	ColorGreen  = "\033[32m"
	ColorYellow = "\033[33m"
	ColorCyan   = "\033[36m"
)

type Field map[string]any

type Logger struct {
	fileWriter io.WriteCloser
	console    io.Writer
}

var globalLogger *Logger

func NewLogger(path string) (*Logger, error) {
	dir := filepath.Dir(path)

	if dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, err
		}
	}

	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, err
	}

	return &Logger{
		fileWriter: file,
		console:    os.Stdout,
	}, nil
}

func InitLogger(path string) error {
	logger, err := NewLogger(path)
	if err != nil {
		return err
	}
	globalLogger = logger
	return nil
}

func CloseLogger() error {
	if globalLogger != nil {
		return globalLogger.fileWriter.Close()
	}
	return nil
}

func BuildLogFilePath(base string) string {
	if base == "" {
		base = "logs/employee-card-admin"
	}

	date := time.Now().Format("2006-01-02")
	return fmt.Sprintf("%s-%s.log", base, date)
}

func (l *Logger) headerColor(level LogLevel) string {
	switch level {
	case DEBUG:
		return ColorCyan
	case INFO:
		return ColorGreen
	case WARN:
		return ColorYellow
	case ERROR:
		return ColorRed
	default:
		return ColorReset
	}
}

func (l *Logger) prettyRaw(v Field) string {
	if len(v) == 0 {
		return "{}"
	}

	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	if err := enc.Encode(v); err != nil {
		return `{ "error": "marshal failed" }`
	}

	return strings.TrimRight(buf.String(), "\n")
}

func (l *Logger) prettyColored(v Field) string {
	raw := l.prettyRaw(v)

	lines := strings.Split(raw, "\n")
	colored := ""

	for _, line := range lines {
		colored += fmt.Sprintf("%s%s%s\n", ColorGreen, line, ColorReset)
	}

	return colored
}

func (l *Logger) write(level LogLevel, component string, fields Field) {
	timestamp := time.Now().Format("2006-01-02 15:04:05")

	hColor := l.headerColor(level)
	header := fmt.Sprintf("[%s] [%s] [%s]", level, component, timestamp)

	fmt.Fprintf(l.console, "%s%s%s\n", hColor, header, ColorReset)
	fmt.Fprint(l.console, l.prettyColored(fields))
	fmt.Fprintf(l.console, "%s------------------------------------------------------------%s\n", hColor, ColorReset)

	if l.fileWriter != nil {
		fmt.Fprintln(l.fileWriter, header)
		fmt.Fprintln(l.fileWriter, l.prettyRaw(fields))
		fmt.Fprintln(l.fileWriter, "------------------------------------------------------------")
	}
}

func Debug(fields Field) {
	if globalLogger != nil {
		globalLogger.write(DEBUG, "", fields)
		return
	}
	log.Printf("[DEBUG] %v", fields)
}

func Info(fields Field) {
	if globalLogger != nil {
		globalLogger.write(INFO, "", fields)
		return
	}
	log.Printf("[INFO] %v", fields)
}

func Warn(fields Field) {
	if globalLogger != nil {
		globalLogger.write(WARN, "", fields)
		return
	}
	log.Printf("[WARN] %v", fields)
}

func Error(fields Field) {
	if globalLogger != nil {
		globalLogger.write(ERROR, "", fields)
		return
	}
	log.Printf("[ERROR] %v", fields)
}

type ComponentLogger struct {
	component string
}

func NewComponentLogger(c string) *ComponentLogger {
	return &ComponentLogger{component: c}
}

func (c *ComponentLogger) write(level LogLevel, fields Field) {
	if globalLogger != nil {
		globalLogger.write(level, c.component, fields)
		return
	}
	log.Printf("[%s][%s] %v", level, c.component, fields)
}

func (c *ComponentLogger) Debug(fields Field) { c.write(DEBUG, fields) }
func (c *ComponentLogger) Info(fields Field)  { c.write(INFO, fields) }
func (c *ComponentLogger) Warn(fields Field)  { c.write(WARN, fields) }
func (c *ComponentLogger) Error(fields Field) { c.write(ERROR, fields) }

package log

import (
	"fmt"
	"os"
)

// LogLevel defines the type for log level definition
// see enum below.
type LogLevel byte

const (
	LEVEL_FATAL LogLevel = iota
	LEVEL_ERROR
	LEVEL_WARNING
	LEVEL_INFO
	LEVEL_DEBUG
	LEVEL_TRACE
)

func LogLevelFromString(level string) LogLevel {
	switch level {
	case "fatal":
		return LEVEL_FATAL
	case "error":
		return LEVEL_ERROR
	case "warning":
		return LEVEL_WARNING
	case "info":
		return LEVEL_INFO
	case "debug":
		return LEVEL_DEBUG
	case "trace":
		return LEVEL_TRACE
	default:
		return LEVEL_INFO
	}
}

// String returns the string representation of the log level.
func (l LogLevel) String() string {
	switch l {
	case LEVEL_FATAL:
		return "FATAL"
	case LEVEL_ERROR:
		return "ERROR"
	case LEVEL_WARNING:
		return "WARNING"
	case LEVEL_INFO:
		return "INFO"
	case LEVEL_DEBUG:
		return "DEBUG"
	case LEVEL_TRACE:
		return "TRACE"
	default:
		return "UNKNOWN"
	}
}

// LogFormat defines the type to define the output format
// see enums below.
type LogFormat int

const (
	FORMAT_JSON LogFormat = iota
	FORMAT_TEXT
)

type LoggerBuilder struct {
	name   *string
	level  LogLevel
	format LogFormat
	file   *os.File
}

func New() *LoggerBuilder {
	return &LoggerBuilder{
		level:  LEVEL_INFO,
		format: FORMAT_TEXT,
	}
}

func (l *LoggerBuilder) Name() *string {
	return l.name
}

func (l *LoggerBuilder) SetName(name string) *LoggerBuilder {
	if l.name == nil {
		l.name = &name
	} else {
		*l.name = fmt.Sprintf("%s|%s", *l.name, name)
	}

	return l
}

func (l *LoggerBuilder) LogLevel() LogLevel {
	return l.level
}

func (l *LoggerBuilder) SetLogLevel(level LogLevel) *LoggerBuilder {
	l.level = level
	return l
}

func (l *LoggerBuilder) LogFormat() LogFormat {
	return l.format
}

func (l *LoggerBuilder) SetLogFormat(format LogFormat) *LoggerBuilder {
	l.format = format
	return l
}

func (l *LoggerBuilder) File() *os.File {
	return l.file
}

func (l *LoggerBuilder) SetFile(file os.File) *LoggerBuilder {
	l.file = &file
	return l
}

func (l *LoggerBuilder) Build() ILogger {
	return &Logger{settings: l}
}

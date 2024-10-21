package log

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

type ILogger interface {
	// Formatted logging
	Fatalf(format string, a ...any)
	Errorf(format string, a ...any)
	Warningf(format string, a ...any)
	Infof(format string, a ...any)
	Debugf(format string, a ...any)
	Tracef(format string, a ...any)
	// Structured logging
	Fatal(msg string, field ...IFlied)
	Error(msg string, field ...IFlied)
	Warning(msg string, field ...IFlied)
	Info(msg string, field ...IFlied)
	Debug(msg string, field ...IFlied)
	Trace(msg string, field ...IFlied)
}

type Logger struct {
	settings *LoggerBuilder
}

func (l *Logger) Fatalf(format string, a ...any) {
	l.log(LEVEL_FATAL, fmt.Sprintf(format, a...))
}

func (l *Logger) Errorf(format string, a ...any) {
	if l.settings.level < LEVEL_ERROR {
		return
	}
	l.log(LEVEL_ERROR, fmt.Sprintf(format, a...))
}

func (l *Logger) Warningf(format string, a ...any) {
	if l.settings.level < LEVEL_WARNING {
		return
	}
	l.log(LEVEL_WARNING, fmt.Sprintf(format, a...))
}

func (l *Logger) Infof(format string, a ...any) {
	if l.settings.level < LEVEL_INFO {
		return
	}
	l.log(LEVEL_INFO, fmt.Sprintf(format, a...))
}

func (l *Logger) Debugf(format string, a ...any) {
	if l.settings.level < LEVEL_DEBUG {
		return
	}
	l.log(LEVEL_DEBUG, fmt.Sprintf(format, a...))
}

func (l *Logger) Tracef(format string, a ...any) {
	if l.settings.level < LEVEL_TRACE {
		return
	}
	l.log(LEVEL_TRACE, fmt.Sprintf(format, a...))
}

func (l *Logger) Fatal(msg string, field ...IFlied) {
	l.log(LEVEL_FATAL, msg, field...)
	os.Exit(1)
}

func (l *Logger) Error(msg string, field ...IFlied) {
	if l.settings.level < LEVEL_ERROR {
		return
	}
	l.log(LEVEL_ERROR, msg, field...)
}

func (l *Logger) Warning(msg string, field ...IFlied) {
	if l.settings.level < LEVEL_WARNING {
		return
	}
	l.log(LEVEL_WARNING, msg, field...)
}

func (l *Logger) Info(msg string, field ...IFlied) {
	if l.settings.level < LEVEL_INFO {
		return
	}
	l.log(LEVEL_INFO, msg, field...)
}

func (l *Logger) Debug(msg string, field ...IFlied) {
	if l.settings.level < LEVEL_DEBUG {
		return
	}
	l.log(LEVEL_DEBUG, msg, field...)
}

func (l *Logger) Trace(msg string, field ...IFlied) {
	if l.settings.level < LEVEL_TRACE {
		return
	}
	l.log(LEVEL_TRACE, msg, field...)
}

func (l *Logger) log(level LogLevel, msg string, field ...IFlied) {
	if l.settings.format == FORMAT_TEXT {
		out := fmt.Sprintf("%s\t%s:\t", time.Now().Format("2006-01-02T15:04:05"), level.String())

		if l.settings.name != nil {
			out = out + *l.settings.name + "\t"
		}

		out = out + msg

		if len(field) > 0 {
			out += "\t"
			for _, f := range field {
				out = out + f.Key() + "=" + f.Value() + ", "
			}
		}

		fmt.Println(out)

		if l.settings.file != nil {
			fmt.Fprintln(l.settings.file, out+"\n")
		}
	} else {
		out := map[string]interface{}{
			"ts":    time.Now(),
			"level": level.String(),
			"msg":   msg,
		}

		if l.settings.name != nil {
			out["caller"] = *l.settings.name
		}

		for _, f := range field {
			out[f.Key()] = f.Value()
		}

		if outJ, err := json.Marshal(out); err == nil {
			fmt.Println(string(outJ))
			if l.settings.file != nil {
				fmt.Fprintln(l.settings.file, string(outJ)+"\n")
			}
		}
	}
}

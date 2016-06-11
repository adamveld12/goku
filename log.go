package goku

import (
	"fmt"
	"log"
	"os"

	"github.com/fatih/color"
	"github.com/hashicorp/logutils"
)

const (
	traceFilter = "TRACE"
	errorFilter = "ERROR"
	fatalFilter = "FATAL"

	debugMinLevel   = logutils.LogLevel(traceFilter)
	releaseMinLevel = logutils.LogLevel(errorFilter)
)

var (
	filter *logutils.LevelFilter
)

func init() {
	minLevel := releaseMinLevel
	if debug {
		minLevel = debugMinLevel
	}

	filter = &logutils.LevelFilter{
		Levels: []logutils.LogLevel{
			traceFilter,
			errorFilter,
			fatalFilter,
		},
		MinLevel: minLevel,
		Writer:   os.Stderr,
	}
}

// NewLog creates a new log.Logger with log filtering
func NewLog(label string) Log {
	return stdErrLogger{
		log.New(filter, label+" ", log.LstdFlags),
		color.New(color.FgBlue).SprintFunc(),
		color.New(color.FgRed).SprintFunc(),
		color.New(color.FgMagenta).SprintFunc(),
	}
}

// Log is a simple interface to facilitate basic logging of errors and debug statements
type Log interface {
	Trace(...interface{})
	Tracef(string, ...interface{})

	Error(...interface{})
	Errorf(string, ...interface{})

	Fatal(...interface{})
	Fatalf(string, ...interface{})
}

type stdErrLogger struct {
	*log.Logger
	traceColorFunc func(...interface{}) string
	errColorFunc   func(...interface{}) string
	fatalColorFunc func(...interface{}) string
}

func (l stdErrLogger) printColorized(label, msg string) {
	output := fmt.Sprintf("%v - %v", label, msg)
	l.Println(output)
}

func (l stdErrLogger) Trace(args ...interface{}) {
	l.printColorized("[TRACE]", l.traceColorFunc(args...))
}
func (l stdErrLogger) Tracef(fmtstr string, args ...interface{}) {
	l.printColorized("[TRACE]", l.traceColorFunc(fmt.Sprintf(fmtstr, args...)))
}

func (l stdErrLogger) Error(args ...interface{}) {
	l.printColorized("[ERROR]", l.errColorFunc(args...))
}

func (l stdErrLogger) Errorf(fmtstr string, args ...interface{}) {
	l.printColorized("[ERROR]", l.errColorFunc(fmt.Sprintf(fmtstr, args...)))
}

func (l stdErrLogger) Fatal(args ...interface{}) {
	l.printColorized("[FATAL]", l.fatalColorFunc(args...))
	panic("FATAL ERROR")
}

func (l stdErrLogger) Fatalf(fmtstr string, args ...interface{}) {
	l.printColorized("[FATAL]", l.fatalColorFunc(fmt.Sprintf(fmtstr, args...)))
	panic("FATAL ERROR")
}

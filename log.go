package goku

import (
	"fmt"
	"log"
	"os"

	"github.com/fatih/color"
	"github.com/hashicorp/logutils"
)

const (
	TraceFilter = "TRACE"
	ErrorFilter = "ERROR"
	FatalFilter = "FATAL"

	debugMinLevel   = logutils.LogLevel(TraceFilter)
	releaseMinLevel = logutils.LogLevel(ErrorFilter)
)

// NewLog creates a new log.Logger with log filtering
func NewLog(label string, debugMode bool) Log {
	minLevel := releaseMinLevel

	if debugMode {
		minLevel = debugMinLevel
	}

	filter := &logutils.LevelFilter{
		Levels: []logutils.LogLevel{
			TraceFilter,
			ErrorFilter,
			FatalFilter,
		},
		MinLevel: minLevel,
		Writer:   os.Stderr,
	}

	return stdErrLogger{
		log.New(filter, label+" ", log.LstdFlags),
		color.New(color.FgBlue).SprintFunc(),
		color.New(color.FgRed).SprintFunc(),
		color.New(color.FgMagenta).SprintFunc(),
	}
}

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

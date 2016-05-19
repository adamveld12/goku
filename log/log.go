package log

import (
	"fmt"
	"io"
	"log"

	"github.com/fatih/color"
	"github.com/hashicorp/logutils"
)

const (
	ErrorFilter = "ERROR"
	WarnFilter  = "WARN"
	DebugFilter = "DEBUG"
	FatalFilter = "FATAL"

	debugMinLevel   = logutils.LogLevel(DebugFilter)
	releaseMinLevel = logutils.LogLevel(WarnFilter)
)

// InitLogging initializes logging levels
func Initialize(debugMode bool, output io.Writer) {

	minLevel := releaseMinLevel

	if debugMode {
		minLevel = debugMinLevel
	}

	filter := &logutils.LevelFilter{
		Levels:   []logutils.LogLevel{"DEBUG", "WARN", "ERROR", "FATAL"},
		MinLevel: minLevel,
		Writer:   output,
	}

	log.SetOutput(filter)
}

func printLog(logType, fmtString string, arguments ...interface{}) {
	colorFunc := color.BlueString

	if logType == ErrorFilter {
		colorFunc = color.RedString
	} else if logType == DebugFilter {
		colorFunc = color.GreenString
	} else if logType == WarnFilter {
		colorFunc = color.YellowString
	} else if logType == FatalFilter {
		colorFunc = color.MagentaString
	} else {
		logType = "INFO"
	}

	log.Println(fmt.Sprintf("[%s] %s", logType, colorFunc(fmtString, arguments...)))
}

func Println(m string) {
	printLog("", "%s\n", m)
}

func Print(m string) {
	printLog("", m)
}

func Fatal(err string) {
	printLog(FatalFilter, "%s", err)
	panic(err)
}

func FatalErr(err error) {
	if err != nil {
		printLog(FatalFilter, "%s", err)
		panic(err)
	}
}

func Debug(output string) {
	printLog(DebugFilter, output)
}

func Debugf(fmtString string, arguments ...interface{}) {
	printLog(DebugFilter, fmtString, arguments...)
}

func DebugErr(err error) {
	if err != nil {
		printLog(DebugFilter, "%s", err.Error())
	}
}

func Err(err error) {
	if err != nil {
		printLog(ErrorFilter, "%s", err.Error())
	}
}

func Error(output string) {
	printLog(ErrorFilter, output)
}

func Errorf(fmtString string, arguments ...interface{}) {
	printLog(ErrorFilter, fmtString, arguments...)
}

func Warn(output string) {
	printLog(WarnFilter, output)
}

func Warnf(fmtString string, arguments ...interface{}) {
	printLog(WarnFilter, fmtString, arguments...)
}

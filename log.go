package main

import (
	"fmt"
	"io"
	"log"

	"github.com/fatih/color"
	"github.com/hashicorp/logutils"
)

const (
	Error = "ERROR"
	Warn  = "WARN"
	Debug = "DEBUG"
	Fatal = "FATAL"
)

const debugMinLevel = logutils.LogLevel("DEBUG")
const releaseMinLevel = logutils.LogLevel("WARN")

// InitLogging initializes logging levels
func InitLogging(debugMode bool, output io.Writer) {

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

func LogFatal(err string) {
	printLog(Fatal, "%s", err)
	panic(err)
}

func LogFatalErr(err error) {
	printLog(Fatal, "%s", err)
	panic(err)
}

func LogDebug(output string) {
	printLog(Debug, output)
}

func LogDebugf(fmtString string, arguments ...interface{}) {
	printLog(Debug, fmtString, arguments...)
}

func LogError(output string) {
	printLog(Error, output)
}

func LogErrorf(fmtString string, arguments ...interface{}) {
	printLog(Error, fmtString, arguments...)
}

func LogWarn(output string) {
	printLog(Warn, output)
}

func LogWarnf(fmtString string, arguments ...interface{}) {
	printLog(Warn, fmtString, arguments...)
}

func printLog(logType, fmtString string, arguments ...interface{}) {
	colorFunc := color.BlueString

	if logType == Error {
		colorFunc = color.RedString
	} else if logType == Debug {
		colorFunc = color.GreenString
	} else if logType == Warn {
		colorFunc = color.YellowString
	} else if logType == Fatal {
		colorFunc = color.MagentaString
	} else {
		logType = "INFO"
	}

	log.Println(fmt.Sprintf("[%s] %s", logType, colorFunc(fmtString, arguments...)))
}

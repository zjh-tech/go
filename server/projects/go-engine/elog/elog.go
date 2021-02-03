package elog

import (
	"fmt"
	"os"
)

var GLogger *Logger

type FuncType func(...interface{})
type ArgType []interface{}

//level: debug 0 info 1 warn 2 error 3
func Init(dir string, level int, initCb FuncType) {
	GLogger = NewLogger(os.Stderr, dir, level)
	GLogger.SetInitCbFunc(initCb)
	GLogger.StartWriterGoroutine()
}

func UnInit(unInitCb FuncType) {
	if GLogger != nil {
		GLogger.SetUnInitCbFunc(unInitCb)
		GLogger.Close()
		GLogger = nil
	}
}

//Debug
func Debug(v ...interface{}) {
	GLogger.AddEvent(LogDebug, fmt.Sprintln(v...), false)
}

func Debugf(format string, v ...interface{}) {
	GLogger.AddEvent(LogDebug, fmt.Sprintf(format, v...), false)
}

func DebugA(v ...interface{}) {
	GLogger.AddEvent(LogDebug, fmt.Sprintln(v...), true)
}

func DebugAf(format string, v ...interface{}) {
	GLogger.AddEvent(LogDebug, fmt.Sprintf(format, v...), true)
}

//Info
func Info(v ...interface{}) {
	GLogger.AddEvent(LogInfo, fmt.Sprintln(v...), false)
}

func Infof(format string, v ...interface{}) {
	GLogger.AddEvent(LogInfo, fmt.Sprintf(format, v...), false)
}

func InfoA(v ...interface{}) {
	GLogger.AddEvent(LogInfo, fmt.Sprintln(v...), true)
}

func InfoAf(format string, v ...interface{}) {
	GLogger.AddEvent(LogInfo, fmt.Sprintf(format, v...), true)
}

//Warn
func Warn(v ...interface{}) {
	GLogger.AddEvent(LogWarn, fmt.Sprintln(v...), false)
}

func Warnf(format string, v ...interface{}) {
	GLogger.AddEvent(LogWarn, fmt.Sprintf(format, v...), false)
}

func WarnA(v ...interface{}) {
	GLogger.AddEvent(LogWarn, fmt.Sprintln(v...), true)
}

func WarnAf(format string, v ...interface{}) {
	GLogger.AddEvent(LogWarn, fmt.Sprintf(format, v...), true)
}

//Error
func Error(v ...interface{}) {
	GLogger.AddEvent(LogError, fmt.Sprintln(v...), false)
}

func Errorf(format string, v ...interface{}) {
	GLogger.AddEvent(LogError, fmt.Sprintf(format, v...), false)
}

func ErrorA(v ...interface{}) {
	GLogger.AddEvent(LogError, fmt.Sprintln(v...), true)
}

func ErrorAf(format string, v ...interface{}) {
	GLogger.AddEvent(LogError, fmt.Sprintf(format, v...), true)
}

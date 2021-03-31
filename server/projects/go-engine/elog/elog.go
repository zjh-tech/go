package elog

import (
	"fmt"
	"os"
)

var GLogger *Logger

type FuncType func(...interface{})
type ArgType []interface{}

//level: debug 0 info 1 warn 2 error 3
func Init(file_dir_prefix string, level int, init_cb FuncType) {
	GLogger = new_logger(os.Stderr, file_dir_prefix, level)
	GLogger.set_init_cb_func(init_cb)
	GLogger.start_writer_goroutine()
}

func UnInit(uninit_cb FuncType) {
	if GLogger != nil {
		GLogger.set_uninit_cb_func(uninit_cb)
		GLogger.close()
		GLogger = nil
	}
}

//Debug
func Debug(v ...interface{}) {
	GLogger.add_event(LogDebug, fmt.Sprintln(v...), false)
}

func Debugf(format string, v ...interface{}) {
	GLogger.add_event(LogDebug, fmt.Sprintf(format, v...), false)
}

func DebugA(v ...interface{}) {
	GLogger.add_event(LogDebug, fmt.Sprintln(v...), true)
}

func DebugAf(format string, v ...interface{}) {
	GLogger.add_event(LogDebug, fmt.Sprintf(format, v...), true)
}

//Info
func Info(v ...interface{}) {
	GLogger.add_event(LogInfo, fmt.Sprintln(v...), false)
}

func Infof(format string, v ...interface{}) {
	GLogger.add_event(LogInfo, fmt.Sprintf(format, v...), false)
}

func InfoA(v ...interface{}) {
	GLogger.add_event(LogInfo, fmt.Sprintln(v...), true)
}

func InfoAf(format string, v ...interface{}) {
	GLogger.add_event(LogInfo, fmt.Sprintf(format, v...), true)
}

//Warn
func Warn(v ...interface{}) {
	GLogger.add_event(LogWarn, fmt.Sprintln(v...), false)
}

func Warnf(format string, v ...interface{}) {
	GLogger.add_event(LogWarn, fmt.Sprintf(format, v...), false)
}

func WarnA(v ...interface{}) {
	GLogger.add_event(LogWarn, fmt.Sprintln(v...), true)
}

func WarnAf(format string, v ...interface{}) {
	GLogger.add_event(LogWarn, fmt.Sprintf(format, v...), true)
}

//Error
func Error(v ...interface{}) {
	GLogger.add_event(LogError, fmt.Sprintln(v...), false)
}

func Errorf(format string, v ...interface{}) {
	GLogger.add_event(LogError, fmt.Sprintf(format, v...), false)
}

func ErrorA(v ...interface{}) {
	GLogger.add_event(LogError, fmt.Sprintln(v...), true)
}

func ErrorAf(format string, v ...interface{}) {
	GLogger.add_event(LogError, fmt.Sprintf(format, v...), true)
}

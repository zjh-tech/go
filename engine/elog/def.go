package elog

import (
	"fmt"
	"sync"
)

const (
	LogDebug = iota
	LogInfo
	LogWarn
	LogError
)

var loglevels = []string{
	"[DEBUG]",
	"[INFO]",
	"[WARN]",
	"[ERROR]",
}

type LogEvent struct {
	level   int
	content string
	file    string
	line    int
}

var GLogEventPool = sync.Pool{
	New: func() interface{} {
		return &LogEvent{}
	},
}

const LogBuffEventSize = 100000
const LogCallDepth = 3

type FuncType func(...interface{})
type ArgType []interface{}

const LogMajorVersion = 1
const LogMinorVersion = 1

type LogVersion struct {
}

func (l *LogVersion) GetVersion() string {
	return fmt.Sprintf("Log Version: %v.%v", LogMajorVersion, LogMinorVersion)
}

var GLogVersion *LogVersion

func init() {
	GLogVersion = &LogVersion{}
}

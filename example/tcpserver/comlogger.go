package main

type ILog interface {
	Debug(v ...interface{})
	Debugf(format string, v ...interface{})
	DebugA(v ...interface{})
	DebugAf(format string, v ...interface{})

	Info(v ...interface{})
	Infof(format string, v ...interface{})
	InfoA(v ...interface{})
	InfoAf(format string, v ...interface{})

	Warn(v ...interface{})
	Warnf(format string, v ...interface{})
	WarnA(v ...interface{})
	WarnAf(format string, v ...interface{})

	Error(v ...interface{})
	Errorf(format string, v ...interface{})
	ErrorA(v ...interface{})
	ErrorAf(format string, v ...interface{})
}

var ELog ILog

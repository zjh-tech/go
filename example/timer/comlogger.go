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

type ComLogger struct {
	logger ILog
}

func NewComLogger() *ComLogger {
	return &ComLogger{}
}

func (c *ComLogger) SetLogger(logger ILog) {
	c.logger = logger
}

func (c *ComLogger) Debug(v ...interface{}) {
	c.logger.Debug(v...)
}

func (c *ComLogger) Debugf(format string, v ...interface{}) {
	c.logger.Debugf(format, v...)
}

func (c *ComLogger) DebugA(v ...interface{}) {
	c.logger.DebugA(v...)
}

func (c *ComLogger) DebugAf(format string, v ...interface{}) {
	c.logger.DebugAf(format, v...)
}

func (c *ComLogger) Info(v ...interface{}) {
	c.logger.Info(v...)
}

func (c *ComLogger) Infof(format string, v ...interface{}) {
	c.logger.Infof(format, v...)
}

func (c *ComLogger) InfoA(v ...interface{}) {
	c.logger.InfoA(v...)
}

func (c *ComLogger) InfoAf(format string, v ...interface{}) {
	c.logger.InfoAf(format, v...)
}

func (c *ComLogger) Warn(v ...interface{}) {
	c.logger.Warn(v...)
}

func (c *ComLogger) Warnf(format string, v ...interface{}) {
	c.logger.Warnf(format, v...)
}

func (c *ComLogger) WarnA(v ...interface{}) {
	c.logger.WarnA(v...)
}

func (c *ComLogger) WarnAf(format string, v ...interface{}) {
	c.logger.WarnAf(format, v...)
}

func (c *ComLogger) Error(v ...interface{}) {
	c.logger.Error(v...)
}

func (c *ComLogger) Errorf(format string, v ...interface{}) {
	c.logger.Errorf(format, v...)
}

func (c *ComLogger) ErrorA(v ...interface{}) {
	c.logger.ErrorA(v...)
}

func (c *ComLogger) ErrorAf(format string, v ...interface{}) {
	c.logger.ErrorAf(format, v...)
}

var ELog *ComLogger

func init() {
	ELog = NewComLogger()
}

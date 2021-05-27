package elog

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"
)

type Logger struct {
	out           io.Writer //os.Stderr -> File
	level         int
	fileDirPrefix string
	events        chan *LogEvent
	buffEvents    chan *LogEvent
	exitChan      chan struct{}
	lastTime      time.Time
	file          *os.File
	buf           bytes.Buffer
	callDepth     int
	closeFlag     bool
	wg            sync.WaitGroup
}

//level: debug 0 info 1 warn 2 error 3
func NewLogger(fileDirPrefix string, level int) *Logger {
	logger := &Logger{
		out:           os.Stderr,
		level:         level,
		fileDirPrefix: fileDirPrefix,
		events:        make(chan *LogEvent),
		buffEvents:    make(chan *LogEvent, LogBuffEventSize),
		exitChan:      make(chan struct{}),
		callDepth:     LogCallDepth,
		closeFlag:     false,
	}

	return logger
}

func (l *Logger) Init() {
	l.startWriterGoroutine()
}

func (l *Logger) UnInit() {
	l.close()
	l.wg.Wait()
}

func (l *Logger) Debug(v ...interface{}) {
	l.addEvent(LogDebug, fmt.Sprint(v...), false)
}

func (l *Logger) Debugf(format string, v ...interface{}) {
	l.addEvent(LogDebug, fmt.Sprintf(format, v...), false)
}

func (l *Logger) DebugA(v ...interface{}) {
	l.addEvent(LogDebug, fmt.Sprint(v...), true)
}

func (l *Logger) DebugAf(format string, v ...interface{}) {
	l.addEvent(LogDebug, fmt.Sprintf(format, v...), true)
}

func (l *Logger) Info(v ...interface{}) {
	l.addEvent(LogInfo, fmt.Sprint(v...), false)
}

func (l *Logger) Infof(format string, v ...interface{}) {
	l.addEvent(LogInfo, fmt.Sprintf(format, v...), false)
}

func (l *Logger) InfoA(v ...interface{}) {
	l.addEvent(LogInfo, fmt.Sprint(v...), true)
}

func (l *Logger) InfoAf(format string, v ...interface{}) {
	l.addEvent(LogInfo, fmt.Sprintf(format, v...), true)
}

func (l *Logger) Warn(v ...interface{}) {
	l.addEvent(LogWarn, fmt.Sprint(v...), false)
}

func (l *Logger) Warnf(format string, v ...interface{}) {
	l.addEvent(LogWarn, fmt.Sprintf(format, v...), false)
}

func (l *Logger) WarnA(v ...interface{}) {
	l.addEvent(LogWarn, fmt.Sprint(v...), true)
}

func (l *Logger) WarnAf(format string, v ...interface{}) {
	l.addEvent(LogWarn, fmt.Sprintf(format, v...), true)
}

func (l *Logger) Error(v ...interface{}) {
	l.addEvent(LogError, fmt.Sprint(v...), false)
}

func (l *Logger) Errorf(format string, v ...interface{}) {
	l.addEvent(LogError, fmt.Sprintf(format, v...), false)
}

func (l *Logger) ErrorA(v ...interface{}) {
	l.addEvent(LogError, fmt.Sprint(v...), true)
}

func (l *Logger) ErrorAf(format string, v ...interface{}) {
	l.addEvent(LogError, fmt.Sprintf(format, v...), true)
}

func (l *Logger) startWriterGoroutine() {
	fmt.Printf("Log Goroutine Start \n")
	l.wg.Add(1)

	go func() {
		defer func() {
			fmt.Printf(" Log Goroutine Exit \n")
			l.wg.Done()
		}()

		exit := false
		for {
			if exit && len(l.buffEvents) == 0 && len(l.events) == 0 {
				//ensure all log write file
				l.out.Write(l.buf.Bytes())
				return
			}
			select {
			case buffEvent := <-l.buffEvents:
				l.outPut(buffEvent.level, buffEvent.content, buffEvent.file, buffEvent.line)

			case event := <-l.events:
				l.outPut(event.level, event.content, event.file, event.line)
			case <-l.exitChan:
				l.exitChan = nil
				exit = true
			}
		}
	}()
}

func (l *Logger) close() {
	l.closeFlag = true
	close(l.exitChan)
}

func (l *Logger) addEvent(level int, content string, async bool) {
	if l.closeFlag {
		return
	}

	if l.level > level {
		return
	}

	_, file, line, _ := runtime.Caller(l.callDepth)
	index := strings.LastIndex(file, "/")
	partFileName := file
	if index != 0 {
		fileLen := len(file)
		partFileName = file[index+1 : fileLen]
	}

	event := &LogEvent{
		level:   level,
		content: content,
		file:    partFileName,
		line:    line,
	}

	if async {
		l.buffEvents <- event
	} else {
		l.events <- event
	}
}

func (l *Logger) outPut(level int, content string, file string, line int) {
	//time zone
	now := time.Now()
	l.ensureFileExist(now)
	l.lastTime = now

	//add time
	nowStr := time.Now().Format("2006-01-02 15:04:05")
	//add file line
	allContent := fmt.Sprintf("%s [%s] [%s:%d] %s", nowStr, loglevels[level], file, line, content)
	l.buf.WriteString(allContent)
	//add \n
	if len(content) > 0 && content[len(content)-1] != '\n' {
		l.buf.WriteByte('\n')
	}

	l.out.Write(l.buf.Bytes())
	l.buf.Reset()

	l.outPutConsole(l.buf.String())
}

func (l *Logger) ensureFileExist(now time.Time) {
	if checkDiffDate(now, l.lastTime) {
		year, month, day := now.Date()
		dir := fmt.Sprintf("%d-%02d-%02d", year, month, day)
		filename := fmt.Sprintf("%d%02d%02d_%02d.log", year, month, day, now.Hour())
		l.createLogFile(dir, filename)
	}
}

func (l *Logger) createLogFile(dir string, filename string) {
	var file *os.File
	fullDir := l.fileDirPrefix + "/" + dir
	_ = createMutiDir(fullDir)
	fullFilePath := fullDir + "/" + filename
	if isExistPath(fullFilePath) {
		file, _ = os.OpenFile(fullFilePath, os.O_APPEND|os.O_RDWR, 0644)
	} else {
		file, _ = os.OpenFile(fullFilePath, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0644)
	}

	if l.file != nil {
		_ = l.file.Close()
		l.file = nil
		l.out = os.Stderr
	}

	l.file = file
	l.out = file
}

func isExistPath(path string) bool {
	var err error
	if _, err = os.Stat(path); err == nil {
		return true
	} else if os.IsExist(err) {
		return true
	}

	return false
}

func createMutiDir(filePath string) error {
	if !isExistPath(filePath) {
		if err := os.MkdirAll(filePath, os.ModePerm); err != nil {
			return err
		}
	}
	return nil
}

func checkDiffDate(now time.Time, last time.Time) bool {
	year, month, day := now.Date()
	hour, _, _ := now.Clock()

	yearl, monthl, dayl := last.Date()
	hourl, _, _ := last.Clock()

	return (year != yearl) || (month != monthl) || (day != dayl) || (hour != hourl)
}

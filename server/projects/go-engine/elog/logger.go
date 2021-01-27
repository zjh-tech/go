package elog

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"
	"time"
)

const (
	LogDebug = iota
	LogInfo
	LogWarn
	LogError
)

var loglevels = []string{
	"DEBUG",
	"INFO",
	"WARN",
	"ERROR",
}

type LogEvent struct {
	Level   int
	Content string
	File    string
	Line    int
}

const LogBuffEventSize = 10000

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
}

func NewLogger(out io.Writer, fileDir string, level int) *Logger {
	logger := &Logger{
		out:           out,
		level:         level,
		fileDirPrefix: fileDir,
		events:        make(chan *LogEvent),
		buffEvents:    make(chan *LogEvent, LogBuffEventSize),
		exitChan:      make(chan struct{}),
		callDepth:     2,
		closeFlag:     false,
	}
	return logger
}

func (l *Logger) StartWriterGoroutine(startCb FuncType, endCb FuncType) {
	fmt.Printf(" Log Goroutine Start \n")

	if startCb != nil {
		startCb()
	}

	go func() {
		defer func() {
			fmt.Printf(" Log Goroutine Exit \n")
			if endCb != nil {
				endCb()
			}
		}()
		exit := false
		for {
			if exit && len(l.buffEvents) == 0 && len(l.events) == 0 {
				//ensure all log write file
				return
			}
			select {
			case buffEvent := <-l.buffEvents:
				l.outPut(buffEvent.Level, buffEvent.Content, buffEvent.File, buffEvent.Line)
			case event := <-l.events:
				l.outPut(event.Level, event.Content, event.File, event.Line)
			case <-l.exitChan:
				l.exitChan = nil
				exit = true
			}
		}
	}()
}

func (l *Logger) Close() {
	l.closeFlag = true
	close(l.exitChan)
}

func (l *Logger) AddEvent(level int, content string, async bool) {
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
		partFileName = file[index+1 : len(file)]
	}

	event := &LogEvent{
		Level:   level,
		Content: content,
		File:    partFileName,
		Line:    line,
	}

	if async == true {
		l.buffEvents <- event
	} else {
		l.events <- event
	}
}

func (l *Logger) outPut(level int, content string, file string, line int) error {
	//time zone
	now := time.Now()
	l.EnsureFileExist(now)
	l.lastTime = now

	l.buf.Reset()
	//add time
	nowStr := time.Now().Format("2006-01-02 15:04:05")
	//add file line
	allContent := fmt.Sprintf("%s [%s] [%s:%d] %s", nowStr, loglevels[level], file, line, content)
	l.buf.WriteString(allContent)
	//add \n
	if len(content) > 0 && content[len(content)-1] != '\n' {
		l.buf.WriteByte('\n')
	}

	_, err := l.out.Write(l.buf.Bytes())
	l.outPutConsole(string(l.buf.Bytes()))
	return err
}

func (l *Logger) EnsureFileExist(now time.Time) {
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
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	return false
}

func createMutiDir(filePath string) error {
	if !isExistPath(filePath) {
		err := os.MkdirAll(filePath, os.ModePerm)
		if err != nil {
			return err
		}
		return err
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

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
	level   int
	content string
	file    string
	line    int
}

const LogBuffEventSize = 10000

type Logger struct {
	out                   io.Writer //os.Stderr -> File
	level                 int
	file_dir_prefix       string
	events                chan *LogEvent
	buff_events           chan *LogEvent
	exit_chan             chan struct{}
	last_time             time.Time
	file                  *os.File
	buf                   bytes.Buffer
	call_depth            int
	close_flag            bool
	init_cb_func          FuncType
	uninit_cb_func        FuncType
	async_flush_max_tick  int64
	next_async_flush_tick int64
}

func new_logger(out io.Writer, file_dir_prefix string, level int) *Logger {
	logger := &Logger{
		out:                  out,
		level:                level,
		file_dir_prefix:      file_dir_prefix,
		events:               make(chan *LogEvent),
		buff_events:          make(chan *LogEvent, LogBuffEventSize),
		exit_chan:            make(chan struct{}),
		call_depth:           2,
		close_flag:           false,
		async_flush_max_tick: 100,
	}

	return logger
}

func (l *Logger) start_writer_goroutine() {
	fmt.Printf(" Log Goroutine Start \n")

	if l.init_cb_func != nil {
		l.init_cb_func()
	}

	go func() {
		defer func() {
			fmt.Printf(" Log Goroutine Exit \n")
			if l.uninit_cb_func != nil {
				l.uninit_cb_func()
			}
		}()

		exit := false
		for {
			if exit && len(l.buff_events) == 0 && len(l.events) == 0 {
				//ensure all log write file
				l.out.Write(l.buf.Bytes())
				return
			}
			select {
			case buffEvent := <-l.buff_events:
				l.out_put(buffEvent.level, buffEvent.content, buffEvent.file, buffEvent.line, true)

			case event := <-l.events:
				l.out_put(event.level, event.content, event.file, event.line, false)
			case <-l.exit_chan:
				l.exit_chan = nil
				exit = true
			}
		}
	}()
}

func (l *Logger) close() {
	l.close_flag = true
	close(l.exit_chan)
}

func (l *Logger) set_init_cb_func(init_cb_func FuncType) {
	l.init_cb_func = init_cb_func
}

func (l *Logger) set_uninit_cb_func(uninit_cb_func FuncType) {
	l.uninit_cb_func = uninit_cb_func
}

func (l *Logger) add_event(level int, content string, async bool) {
	if l.close_flag {
		return
	}

	if l.level > level {
		return
	}

	_, file, line, _ := runtime.Caller(l.call_depth)
	index := strings.LastIndex(file, "/")
	partFileName := file
	if index != 0 {
		partFileName = file[index+1 : len(file)]
	}

	event := &LogEvent{
		level:   level,
		content: content,
		file:    partFileName,
		line:    line,
	}

	if async == true {
		l.buff_events <- event
	} else {
		l.events <- event
	}
}

func get_mill_second() int64 {
	return time.Now().UnixNano() / 1e6
}

func (l *Logger) out_put(level int, content string, file string, line int, async_flag bool) {
	//time zone
	now := time.Now()
	l.ensure_file_fxist(now)
	l.last_time = now

	//add time
	nowStr := time.Now().Format("2006-01-02 15:04:05")
	//add file line
	allContent := fmt.Sprintf("%s [%s] [%s:%d] %s", nowStr, loglevels[level], file, line, content)
	l.buf.WriteString(allContent)
	//add \n
	if len(content) > 0 && content[len(content)-1] != '\n' {
		l.buf.WriteByte('\n')
	}

	if async_flag {
		if l.next_async_flush_tick < get_mill_second() {
			l.out.Write(l.buf.Bytes())
			l.buf.Reset()
		}
	} else {
		l.out.Write(l.buf.Bytes())
		l.buf.Reset()
	}

	l.out_put_console(string(l.buf.Bytes()))
}

func (l *Logger) ensure_file_fxist(now time.Time) {
	if check_diff_date(now, l.last_time) {
		year, month, day := now.Date()
		dir := fmt.Sprintf("%d-%02d-%02d", year, month, day)
		filename := fmt.Sprintf("%d%02d%02d_%02d.log", year, month, day, now.Hour())
		l.create_log_file(dir, filename)
	}
}

func (l *Logger) create_log_file(dir string, filename string) {
	var file *os.File
	fullDir := l.file_dir_prefix + "/" + dir
	_ = create_muti_dir(fullDir)
	fullFilePath := fullDir + "/" + filename
	if is_exist_path(fullFilePath) {
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

func is_exist_path(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	return false
}

func create_muti_dir(file_path string) error {
	if !is_exist_path(file_path) {
		err := os.MkdirAll(file_path, os.ModePerm)
		if err != nil {
			return err
		}
		return err
	}
	return nil
}

func check_diff_date(now time.Time, last time.Time) bool {
	year, month, day := now.Date()
	hour, _, _ := now.Clock()

	yearl, monthl, dayl := last.Date()
	hourl, _, _ := last.Clock()

	return (year != yearl) || (month != monthl) || (day != dayl) || (hour != hourl)
}

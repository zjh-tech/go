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
	out             io.Writer //os.Stderr -> File
	level           int
	file_dir_prefix string
	events          chan *LogEvent
	buff_events     chan *LogEvent
	exit_chan       chan struct{}
	last_time       time.Time
	file            *os.File
	buf             bytes.Buffer
	call_depth      int
	close_flag      bool
	wg              sync.WaitGroup
}

//level: debug 0 info 1 warn 2 error 3
func NewLogger(file_dir_prefix string, level int) *Logger {
	logger := &Logger{
		out:             os.Stderr,
		level:           level,
		file_dir_prefix: file_dir_prefix,
		events:          make(chan *LogEvent),
		buff_events:     make(chan *LogEvent, log_buff_event_size),
		exit_chan:       make(chan struct{}),
		call_depth:      log_call_depth,
		close_flag:      false,
	}

	return logger
}

func (l *Logger) Init() {
	l.start_writer_goroutine()
}

func (l *Logger) UnInit() {
	l.close()
	l.wg.Wait()
}

func (l *Logger) Debug(v ...interface{}) {
	l.add_event(LogDebug, fmt.Sprint(v...), false)
}

func (l *Logger) Debugf(format string, v ...interface{}) {
	l.add_event(LogDebug, fmt.Sprintf(format, v...), false)
}

func (l *Logger) DebugA(v ...interface{}) {
	l.add_event(LogDebug, fmt.Sprint(v...), true)
}

func (l *Logger) DebugAf(format string, v ...interface{}) {
	l.add_event(LogDebug, fmt.Sprintf(format, v...), true)
}

func (l *Logger) Info(v ...interface{}) {
	l.add_event(LogInfo, fmt.Sprint(v...), false)
}

func (l *Logger) Infof(format string, v ...interface{}) {
	l.add_event(LogInfo, fmt.Sprintf(format, v...), false)
}

func (l *Logger) InfoA(v ...interface{}) {
	l.add_event(LogInfo, fmt.Sprint(v...), true)
}

func (l *Logger) InfoAf(format string, v ...interface{}) {
	l.add_event(LogInfo, fmt.Sprintf(format, v...), true)
}

func (l *Logger) Warn(v ...interface{}) {
	l.add_event(LogWarn, fmt.Sprint(v...), false)
}

func (l *Logger) Warnf(format string, v ...interface{}) {
	l.add_event(LogWarn, fmt.Sprintf(format, v...), false)
}

func (l *Logger) WarnA(v ...interface{}) {
	l.add_event(LogWarn, fmt.Sprint(v...), true)
}

func (l *Logger) WarnAf(format string, v ...interface{}) {
	l.add_event(LogWarn, fmt.Sprintf(format, v...), true)
}

func (l *Logger) Error(v ...interface{}) {
	l.add_event(LogError, fmt.Sprint(v...), false)
}

func (l *Logger) Errorf(format string, v ...interface{}) {
	l.add_event(LogError, fmt.Sprintf(format, v...), false)
}

func (l *Logger) ErrorA(v ...interface{}) {
	l.add_event(LogError, fmt.Sprint(v...), true)
}

func (l *Logger) ErrorAf(format string, v ...interface{}) {
	l.add_event(LogError, fmt.Sprintf(format, v...), true)
}

func (l *Logger) start_writer_goroutine() {
	fmt.Printf("Log Goroutine Start \n")
	l.wg.Add(1)

	go func() {
		defer func() {
			fmt.Printf(" Log Goroutine Exit \n")
			l.wg.Done()
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
				l.out_put(buffEvent.level, buffEvent.content, buffEvent.file, buffEvent.line)

			case event := <-l.events:
				l.out_put(event.level, event.content, event.file, event.line)
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

func (l *Logger) add_event(level int, content string, async bool) {
	if l.close_flag {
		return
	}

	if l.level > level {
		return
	}

	_, file, line, _ := runtime.Caller(l.call_depth)
	index := strings.LastIndex(file, "/")
	part_file_name := file
	if index != 0 {
		part_file_name = file[index+1 : len(file)]
	}

	event := &LogEvent{
		level:   level,
		content: content,
		file:    part_file_name,
		line:    line,
	}

	if async == true {
		l.buff_events <- event
	} else {
		l.events <- event
	}
}

func (l *Logger) out_put(level int, content string, file string, line int) {
	//time zone
	now := time.Now()
	l.ensure_file_fxist(now)
	l.last_time = now

	//add time
	now_str := time.Now().Format("2006-01-02 15:04:05")
	//add file line
	all_content := fmt.Sprintf("%s [%s] [%s:%d] %s", now_str, loglevels[level], file, line, content)
	l.buf.WriteString(all_content)
	//add \n
	if len(content) > 0 && content[len(content)-1] != '\n' {
		l.buf.WriteByte('\n')
	}

	l.out.Write(l.buf.Bytes())
	l.buf.Reset()

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
	full_dir := l.file_dir_prefix + "/" + dir
	_ = create_muti_dir(full_dir)
	full_file_path := full_dir + "/" + filename
	if is_exist_path(full_file_path) {
		file, _ = os.OpenFile(full_file_path, os.O_APPEND|os.O_RDWR, 0644)
	} else {
		file, _ = os.OpenFile(full_file_path, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0644)
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
	var err error
	if _, err = os.Stat(path); err == nil {
		return true
	} else if os.IsExist(err) {
		return true
	}

	return false
}

func create_muti_dir(file_path string) error {
	if !is_exist_path(file_path) {
		if err := os.MkdirAll(file_path, os.ModePerm); err != nil {
			return err
		}
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

package frame

import (
	"bytes"
	"fmt"
	"runtime"
	"strings"
	"time"

	"github.com/zjh-tech/go-frame/base/convert"
)

func getMillsecond() int64 {
	return time.Now().UnixNano() / 1e6
}

const (
	TIME_METER_STAMP_MAX_SIZE int = 20
)

type TimeMeter struct {
	limitMillSec   int64
	index          int
	stampTicks     []int64
	buf            bytes.Buffer
	startStampTick int64
}

func NewTimeMeter(limitMillSec int64) *TimeMeter {
	return &TimeMeter{
		limitMillSec: limitMillSec,
		stampTicks:   make([]int64, TIME_METER_STAMP_MAX_SIZE),
	}
}

func (t *TimeMeter) Stamp() {
	if t.index < TIME_METER_STAMP_MAX_SIZE {
		t.stampTicks[t.index] = getMillsecond()
		t.index++
	}
}

func (t *TimeMeter) Clear() {
	t.index = 0
	t.startStampTick = getMillsecond()
	t.buf.Reset()
}

func (t *TimeMeter) CheckOut() {
	t.Stamp()

	if t.index != 0 && (t.stampTicks[t.index-1]-t.startStampTick >= t.limitMillSec) {
		_, file, line, _ := runtime.Caller(2)
		index := strings.LastIndex(file, "/")
		partFileName := file
		if index != 0 {
			partFileName = file[index+1 : len(file)]
		}

		t.buf.WriteString(fmt.Sprintf("[%s:%d]", partFileName, line))
		for i := 0; i < t.index; i++ {
			if i != 0 {
				t.buf.WriteString("-")
			}
			t.buf.WriteString("[")
			t.buf.WriteString(convert.Int642Str(t.stampTicks[i] - t.startStampTick))
			t.buf.WriteString("]")
		}

		ELog.WarnAf("[TimeMeter] TimeOut: %v", t.buf.String())
	}
}

package main

import (
	"projects/frame"
	"projects/go-engine/elog"
	"time"
)

func main() {
	logger := elog.NewLogger("./log", 0)
	logger.Init()
	ELog.SetLogger(logger)

	test_day_delay()
	test_week_delay()
	test_month_delay()

	for {
		time.Sleep(1 * time.Millisecond)
	}
}

func test_day_delay() {
	cur_time := time.Now()
	hour := uint64(cur_time.Hour())
	min := uint64(cur_time.Minute())
	sec := uint64(cur_time.Second())

	diff := uint64(1)
	flag1 := false
	hour1 := hour
	if hour+diff < 24 && !flag1 {
		hour1 = hour + diff
		flag1 = true
	}
	min1 := min
	if min+diff < 60 && !flag1 {
		min1 = min + diff
		flag1 = true
	}
	sec1 := sec
	if sec+diff < 60 && !flag1 {
		sec1 = sec + diff
		flag1 = true
	}

	delay := frame.GetNextDelayEDayHMS(hour1, min1, sec1)
	ELog.InfoAf("[Day Delay] [+] Hour=%v,Min=%v,Sec=%v Delay=%v", hour1-hour, min1-min, sec1-sec, delay)

	flag2 := false
	hour2 := hour
	if hour-diff > 0 && !flag2 {
		hour2 = hour - diff
		flag2 = true
	}
	min2 := min
	if min-diff > 0 && !flag2 {
		min2 = min - diff
		flag2 = true
	}
	sec2 := sec
	if sec-diff > 0 && !flag2 {
		sec2 = sec - diff
		flag2 = true
	}

	delay2 := frame.GetNextDelayEDayHMS(hour2, min2, sec2)
	ELog.InfoAf("[Day Delay] [-] Hour=%v,Min=%v,Sec=%v Delay=%v", hour-hour2, min-min2, sec-sec2, delay2)
}

func test_week_delay() {

}

func test_month_delay() {

}

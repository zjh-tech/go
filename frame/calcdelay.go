package frame

import (
	"math"
	"time"

	"github.com/zjh-tech/go-frame/base/util"
)

const (
	NOVALID_DELAY_MILL_MARCO uint64 = math.MaxInt64
)

const (
	PER_MILL_SECOND = uint64(1)
	PER_SECOND      = PER_MILL_SECOND * 1000
	PER_MINUTE      = PER_SECOND * 60
	PER_HOUR        = PER_MINUTE * 60
	PER_DAY         = PER_HOUR * 24
	PER_WEEK        = PER_DAY * 7
)

//condition  60 % min == 0
//triggle time: day.hour.min1.sec ==> day.hour.(min1 + min).sec
//func GetNextDelayEMin(min uint64) uint64 {
//	if (min < 0 || min >= 60) || (60%min != 0) {
//		elog.ErrorAf("GetNextDelayEMin Min=%v", min)
//		return NOVALID_DELAY_MILL_MARCO
//	}
//	cur_time := time.Now()
//	cur_min := uint64(cur_time.Minute())
//	cur_sec := uint64(cur_time.Second())
//
//	remain_time := PER_HOUR - (PER_MINUTE*cur_min + PER_SECOND*cur_sec)
//	delay := remain_time % (PER_MINUTE * min)
//	return delay
//}

//condition 24 % hour == 0
//triggle time: day.hour1.min.sec ==> day.(hour1+hour).min.sec...
//func GetNextDelayDEHMS(hour uint64, min uint64, sec uint64) uint64 {
//	if (hour < 0 || hour >= 24) || (min < 0 || min >= 60) || (sec < 0 || sec >= 60) || (24%hour) != 0 {
//		elog.ErrorAf("GetNearestDelayEveryDay Hour=%v Min=%v Sec=%v", hour, min, sec)
//		return NOVALID_DELAY_MILL_MARCO
//	}
//
//	cur_time := time.Now()
//	cur_hour := uint64(cur_time.Hour())
//	cur_min := uint64(cur_time.Minute())
//	cur_sec := uint64(cur_time.Second())
//
//	ret_sec := uint64(0)
//
//	diff_time := hour*PER_HOUR + min*PER_MINUTE + sec*PER_SECOND
//	cur_diff_time := cur_hour%hour*PER_HOUR + cur_min*PER_MINUTE + cur_sec*PER_SECOND
//	if cur_diff_time < diff_time {
//		ret_sec = diff_time - cur_diff_time
//	} else {
//		ret_sec = diff_time + diff_time - cur_diff_time
//	}
//	return ret_sec
//}

//------------------------------日历 api 开始---------------------------------------
//triggle time: day.hour.min.sec ==> day+1.hour.min.sec ...
func GetNextDelayEDayHMS(hour uint64, min uint64, sec uint64) uint64 {
	if (hour < 0 || hour >= 24) || (min < 0 || min >= 60) || (sec < 0 || sec >= 60) {
		ELog.ErrorAf("GetNextDelayEDayHMS Hour=%v Min=%v Sec=%v", hour, min, sec)
		return NOVALID_DELAY_MILL_MARCO
	}

	cur_time := time.Now()
	cur_hour := uint64(cur_time.Hour())
	cur_min := uint64(cur_time.Minute())
	cur_sec := uint64(cur_time.Second())

	ret_sec := uint64(0)
	diff_time := hour*PER_HOUR + min*PER_MINUTE + sec*PER_SECOND
	cur_diff_time := cur_hour*PER_HOUR + cur_min*PER_MINUTE + cur_sec*PER_SECOND

	if cur_diff_time < diff_time {
		ret_sec = diff_time - cur_diff_time
	} else {
		ret_sec = PER_DAY + diff_time - cur_diff_time
	}

	return ret_sec
}

//triggle time: wday.hour.min.sec ==> wday+7.hour.min.sec ...
func GetNextDelayEWeekHMS(wday uint64, hour uint64, min uint64, sec uint64) uint64 {
	if (wday < 0 || wday > 6) || (hour < 0 || hour >= 24) || (min < 0 || min >= 60) || (sec < 0 || sec >= 60) {
		ELog.ErrorAf("GetNextDelayEWeekHMS Wday=%v Hour=%v Min=%v Sec=%v", wday, hour, min, sec)
		return NOVALID_DELAY_MILL_MARCO
	}

	cur_time := time.Now()
	now := uint64(cur_time.UnixNano() / 1e6)
	cur_wday := uint64(cur_time.Weekday())

	triggle_time := uint64(time.Date(cur_time.Year(), cur_time.Month(), cur_time.Day(), int(hour), int(min), int(sec), 0, time.Local).UnixNano() / 1e6)
	if cur_wday < wday {
		triggle_time = triggle_time - (wday-cur_wday)*PER_DAY
	} else {
		triggle_time = triggle_time + (wday+7-cur_wday)*PER_DAY
	}

	ret_sec := uint64(0)
	if triggle_time > now {
		ret_sec = triggle_time - now
	} else {
		ret_sec = triggle_time + PER_WEEK - now
	}
	return ret_sec
}

//triggle time: month.day.hour.min.sec ==> month+1.day.hour.min.sec ...
func GetNextDelayEMonthDHMS(day uint64, hour uint64, min uint64, sec uint64) uint64 {
	if (day < 1 || day > 28) || (hour < 0 || hour >= 24) || (min < 0 || min >= 60) || (sec < 0 || sec >= 60) {
		ELog.ErrorAf("GetNextDelayEMonthDHMS day=%v Hour=%v Min=%v,Sec=%v", day, hour, min, sec)
		return NOVALID_DELAY_MILL_MARCO
	}

	cur_time := time.Now()
	now := uint64(cur_time.UnixNano() / 1e6)

	triggle_time := uint64(time.Date(cur_time.Year(), cur_time.Month(), int(day), int(hour), int(min), int(sec), 0, time.Local).UnixNano() / 1e6)
	ret_sec := uint64(0)
	if triggle_time > now {
		ret_sec = triggle_time - now
	} else {
		_, total_day := util.GetTotalDayByMonth(uint64(cur_time.Year()), uint64(cur_time.Month()+1))
		ret_sec = triggle_time + PER_DAY*total_day - now
	}

	return ret_sec
}

//------------------------------日历 api 结束---------------------------------------

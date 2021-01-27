package frame

import (
	"math"
	"projects/go-engine/elog"
	"projects/util"
	"time"
)

const (
	NOVALID_DELAY_MILL_MARCO int64 = math.MaxInt64
)

//everyhour  min   5 ==> triggle time (1.5) (2.5) (3.5)...
//condition  60 % min == 0
func GetNearestDelayEveryMin(min int64) int64 {
	if (min < 0) || (min >= 60) || (60%min != 0) {
		elog.ErrorAf("GetNearestDelayEveryMin Min=%v", min)
		return NOVALID_DELAY_MILL_MARCO
	}
	cur_time := time.Now()
	cur_min := int64(cur_time.Minute())
	cur_sec := int64(cur_time.Second())

	remain_time := int64(time.Hour) - (int64(time.Minute)*cur_min + int64(time.Second)*cur_sec)
	delay := remain_time % (int64(time.Minute) * min)
	return delay
}

//everyday  hour = 0 min = 0 ==> triggle time(1.0.0) (2.0.0) (3.0.0)...
//hour(1-23) min(0-59)
//condition  24 % hour == 0
func GetNearestDelayEveryDay(hour int64, min int64) int64 {
	if hour < 0 || hour >= 24 || min < 0 || min >= 60 || (24%hour) != 0 {
		elog.ErrorAf("GetNearestDelayEveryDay Hour=%v Min=%v ", hour, min)
		return NOVALID_DELAY_MILL_MARCO
	}

	cur_time := time.Now()
	cur_hour := int64(cur_time.Hour())
	cur_min := int64(cur_time.Minute())
	cur_sec := int64(cur_time.Second())
	ret_nano_sec := int64(0)
	if (int64(cur_time.Hour())%hour == 0) && (int64(cur_time.Minute()) < min) {
		ret_nano_sec = (min-cur_min)*int64(time.Minute) - cur_sec*int64(time.Second)
	} else {
		next_diff_time := hour*int64(time.Hour) + min*int64(time.Minute)
		cur_diff_time := (cur_hour)%hour*int64(time.Hour) + cur_min*int64(time.Minute) + cur_sec*int64(time.Second)
		ret_nano_sec = next_diff_time - cur_diff_time
	}

	return ret_nano_sec / 1e6
}

//everyweek  hour = 0 min = 0
//wday(1-7)hour(0-23)min(0-59)
func GetNearestDelayEveryWeek(wday int64, hour int64, min int64) int64 {
	if wday < 1 || wday > 7 || hour < 0 || hour >= 24 || min < 0 || min >= 60 {
		elog.ErrorAf("GetNearestDelayEveryWeek Wday=%v Hour=%v Min=%v ", wday, hour, min)
		return NOVALID_DELAY_MILL_MARCO
	}

	cur_time := time.Now()
	now := cur_time.UnixNano()
	cur_wday := int64(cur_time.Weekday())
	cur_hour := int64(cur_time.Hour())
	cur_min := int64(cur_time.Minute())
	cur_sec := int64(cur_time.Second())
	cur_week_triggle_time := get_current_week_triggle_time(cur_wday, cur_hour, cur_min, cur_sec)
	if cur_week_triggle_time < now {
		return (cur_week_triggle_time - now) / 1e6
	} else {
		return (cur_week_triggle_time + (int64(time.Hour) * 24 * 7) + -now) / 1e6
	}
}

//everymonth day = 1 hour = 0 min = 0 ==> triggle time(1.0.0.0) (2.0.0.0)
//day(1-31)hour(0-23)min(0-59)
func GetNearestDelayEveryMonth(day int64, hour int64, min int64) int64 {
	if day < 1 || day > 28 || hour < 0 || hour > 24 || min < 0 || min > 59 {
		elog.ErrorAf("GetNearestDelayEveryMonth day=%v Hour=%v Min=%v ", day, hour, min)
		return NOVALID_DELAY_MILL_MARCO
	}

	cur_time := time.Now()
	now := cur_time.UnixNano()
	cur_month_triggle_time := get_current_month_triggle_time(day, hour, min, 0)
	if cur_month_triggle_time > now {
		return (cur_month_triggle_time - now) / 1e6
	} else {
		next_month := (cur_time.Month() + 1) % 12
		if next_month == 0 {
			next_month = 12
		}

		_, total_day := util.GetTotalDayByMonth(int64(cur_time.Year()), int64(next_month))
		next_month_triggle_time := cur_month_triggle_time + (int64(time.Hour) * 24 * total_day)
		return (next_month_triggle_time - now) / 1e6
	}
}

func get_current_week_triggle_time(wday int64, hour int64, min int64, sec int64) int64 {
	cur_time := time.Now()
	triggle_time := time.Date(cur_time.Year(), cur_time.Month(), cur_time.Day(), int(hour), int(min), int(sec), 0, time.Local)
	cur_wday := int64(cur_time.Weekday())
	ret_nano_sec := triggle_time.UnixNano()
	if int64(cur_time.Weekday()) > wday {
		temp_nano_sec := (cur_wday - wday) * (int64(time.Hour) * 24)
		ret_nano_sec -= temp_nano_sec
	} else {
		temp_nano_sec := (wday - cur_wday) * (int64(time.Hour) * 24)
		ret_nano_sec += temp_nano_sec
	}

	return ret_nano_sec
}

func get_current_month_triggle_time(day int64, hour int64, min int64, sec int64) int64 {
	cur_time := time.Now()
	triggle_time := time.Date(cur_time.Year(), cur_time.Month(), int(day), int(hour), int(min), int(sec), 0, time.Local)
	return (triggle_time.UnixNano())
}

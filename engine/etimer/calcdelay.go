package etimer

import (
	"math"
	"time"

	"github.com/zjh-tech/go-frame/base/util"
)

const (
	NovalidDelayMillMarco uint64 = math.MaxInt64
)

const (
	PerMillSecond = uint64(1)
	PerSecond     = PerMillSecond * 1000
	PerMinute     = PerSecond * 60
	PerHour       = PerMinute * 60
	PerDay        = PerHour * 24
	PerWeek       = PerDay * 7
)

//condition  60 % min == 0
//triggle time: day.hour.min1.sec ==> day.hour.(min1 + min).sec
//func GetNextDelayEMin(min uint64) uint64 {
//	if (min < 0 || min >= 60) || (60%min != 0) {
//		elog.ErrorAf("GetNextDelayEMin Min=%v", min)
//		return NovalidDelayMillMarco
//	}
//	curTime := time.Now()
//	curMin := uint64(curTime.Minute())
//	curSec := uint64(curTime.Second())
//
//	remainTime := PerHour - (PerMinute*curMin + PerSecond*curSec)
//	delay := remainTime % (PerMinute * min)
//	return delay
//}

//condition 24 % hour == 0
//triggle time: day.hour1.min.sec ==> day.(hour1+hour).min.sec...
//func GetNextDelayDEHMS(hour uint64, min uint64, sec uint64) uint64 {
//	if (hour < 0 || hour >= 24) || (min < 0 || min >= 60) || (sec < 0 || sec >= 60) || (24%hour) != 0 {
//		elog.ErrorAf("GetNearestDelayEveryDay Hour=%v Min=%v Sec=%v", hour, min, sec)
//		return NovalidDelayMillMarco
//	}
//
//	curTime := time.Now()
//	curHour := uint64(curTime.Hour())
//	curMin := uint64(curTime.Minute())
//	curSec := uint64(curTime.Second())
//
//	retSec := uint64(0)
//
//	diffTime := hour*PerHour + min*PerMinute + sec*PerSecond
//	curDiffTime := curHour%hour*PerHour + curMin*PerMinute + curSec*PerSecond
//	if curDiffTime < diffTime {
//		retSec = diffTime - curDiffTime
//	} else {
//		retSec = diffTime + diffTime - curDiffTime
//	}
//	return retSec
//}

//------------------------------日历 api 开始---------------------------------------
//triggle time: day.hour.min.sec ==> day+1.hour.min.sec ...
func GetNextDelayEDayHMS(hour uint64, min uint64, sec uint64) uint64 {
	if hour >= 24 || min >= 60 || sec >= 60 {
		ELog.ErrorAf("GetNextDelayEDayHMS Hour=%v Min=%v Sec=%v", hour, min, sec)
		return NovalidDelayMillMarco
	}

	curTime := time.Now()
	curHour := uint64(curTime.Hour())
	curMin := uint64(curTime.Minute())
	curSec := uint64(curTime.Second())

	retSec := uint64(0)
	diffTime := hour*PerHour + min*PerMinute + sec*PerSecond
	curDiffTime := curHour*PerHour + curMin*PerMinute + curSec*PerSecond

	if curDiffTime < diffTime {
		retSec = diffTime - curDiffTime
	} else {
		retSec = PerDay + diffTime - curDiffTime
	}

	return retSec
}

//triggle time: wday.hour.min.sec ==> wday+7.hour.min.sec ...
func GetNextDelayEWeekHMS(wday uint64, hour uint64, min uint64, sec uint64) uint64 {
	if wday > 6 || hour >= 24 || min >= 60 || sec >= 60 {
		ELog.ErrorAf("GetNextDelayEWeekHMS Wday=%v Hour=%v Min=%v Sec=%v", wday, hour, min, sec)
		return NovalidDelayMillMarco
	}

	curTime := time.Now()
	now := uint64(curTime.UnixNano() / 1e6)
	curWday := uint64(curTime.Weekday())

	triggleTime := uint64(time.Date(curTime.Year(), curTime.Month(), curTime.Day(), int(hour), int(min), int(sec), 0, time.Local).UnixNano() / 1e6)
	if curWday < wday {
		triggleTime = triggleTime - (wday-curWday)*PerDay
	} else {
		triggleTime = triggleTime + (wday+7-curWday)*PerDay
	}

	retSec := uint64(0)
	if triggleTime > now {
		retSec = triggleTime - now
	} else {
		retSec = triggleTime + PerWeek - now
	}
	return retSec
}

//triggle time: month.day.hour.min.sec ==> month+1.day.hour.min.sec ...
func GetNextDelayEMonthDHMS(day uint64, hour uint64, min uint64, sec uint64) uint64 {
	if (day < 1 || day > 28) || hour >= 24 || min >= 60 || sec >= 60 {
		ELog.ErrorAf("GetNextDelayEMonthDHMS day=%v Hour=%v Min=%v,Sec=%v", day, hour, min, sec)
		return NovalidDelayMillMarco
	}

	curTime := time.Now()
	now := uint64(curTime.UnixNano() / 1e6)

	triggleTime := uint64(time.Date(curTime.Year(), curTime.Month(), int(day), int(hour), int(min), int(sec), 0, time.Local).UnixNano() / 1e6)
	retSec := uint64(0)
	if triggleTime > now {
		retSec = triggleTime - now
	} else {
		_, total_day := util.GetTotalDayByMonth(uint64(curTime.Year()), uint64(curTime.Month()+1))
		retSec = triggleTime + PerDay*total_day - now
	}

	return retSec
}

//------------------------------日历 api 结束---------------------------------------

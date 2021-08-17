package enet

//游戏不采用速率限制器rate.NewLimiter,Web能接受丢包
type OverLoadModule struct {
	count             int64
	lastCheckTime     int64
	limit             int64
	checkIntervalTime int64
}

func NewOverLoadModule(intervalTime int64, limit int64) *OverLoadModule {
	return &OverLoadModule{
		count:             0,
		lastCheckTime:     getSecond(),
		limit:             limit,
		checkIntervalTime: intervalTime,
	}
}

func (o *OverLoadModule) AddCount() {
	o.count++
}

func (o *OverLoadModule) IsOverLoad() bool {
	now := getMillsecond()
	if (o.lastCheckTime + o.checkIntervalTime) > now {
		return false
	}

	if o.count/o.checkIntervalTime > o.limit {
		return true
	}

	o.lastCheckTime = now
	o.count = 0
	return false

}

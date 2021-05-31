package etimer

type FuncType func(...interface{})
type ArgType []interface{}

type ITimerMgr interface {
	Update(loopCount int) bool
	CreateSlotTimer(eid uint32, delay uint64, repeat bool, cb FuncType, args ArgType, r *TimerRegister) *Timer
	AddSlotTimer(timer *Timer)
}

type ITimerRegister interface {
	AddOnceTimer(id uint32, delay uint64, cb FuncType, args ArgType, replace bool)
	AddRepeatTimer(id uint32, delay uint64, cb FuncType, args ArgType, replace bool)
	HasTimer(id uint32) bool
	KillTimer(id uint32)
	KillAllTimer()
	GetRemainTime(id uint32) (bool, uint64)
}

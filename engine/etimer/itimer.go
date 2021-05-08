package etimer

type FuncType func(...interface{})
type ArgType []interface{}

type ITimerMgr interface {
	Update(loop_count int)
}

type ITimerRegister interface {
	AddOnceTimer(id uint32, delay uint64, desc string, cb FuncType, args ArgType, replace bool)
	AddRepeatTimer(id uint32, delay uint64, desc string, cb FuncType, args ArgType, replace bool)
	HasTimer(id uint32) bool
	KillTimer(id uint32)
	KillAllTimer()
	GetRemainTime(id uint32) (bool, uint64)
}

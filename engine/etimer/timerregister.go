package etimer

import (
	"runtime"
)

type TimerRegister struct {
	registerMap map[uint32]*Timer
	mgr         ITimerMgr
}

func CleanTimer(t *TimerRegister) {
	t.KillAllTimer()
}

func NewTimerRegister(mgr ITimerMgr) *TimerRegister {
	tr := &TimerRegister{
		registerMap: make(map[uint32]*Timer),
		mgr:         mgr,
	}
	runtime.SetFinalizer(tr, CleanTimer)
	return tr
}

func (t *TimerRegister) AddOnceTimer(id uint32, delay uint64, cb FuncType, args ArgType, replace bool) {
	t.addTimer(id, delay, false, cb, args)
}

func (t *TimerRegister) AddRepeatTimer(id uint32, delay uint64, cb FuncType, args ArgType, replace bool) {
	t.addTimer(id, delay, true, cb, args)
}

func (t *TimerRegister) HasTimer(id uint32) bool {
	_, ok := t.registerMap[id]
	return ok
}

func (t *TimerRegister) GetRemainTime(id uint32) (bool, uint64) {
	timer, ok := t.registerMap[id]
	if !ok {
		return false, uint64(0)
	}

	return true, timer.getRemainTime()
}

func (t *TimerRegister) KillTimer(id uint32) {
	timer, ok := t.registerMap[id]
	if ok {
		timer.Kill()
		delete(t.registerMap, id)
	}
}

func (t *TimerRegister) KillAllTimer() {
	for _, timer := range t.registerMap {
		timer.Kill()
	}
	t.registerMap = make(map[uint32]*Timer)
}

func (t *TimerRegister) RemoveTimer(info *Timer) {
	timer, ok := t.registerMap[info.eid]
	if ok {
		if timer.uid == info.uid {
			delete(t.registerMap, timer.eid)
		}
	}
}

func (t *TimerRegister) addTimer(id uint32, delay uint64, repeat bool, cb FuncType, args ArgType) bool {
	if delay == NovalidDelayMill {
		return false
	}

	exist := t.HasTimer(id)
	if exist {
		return true
	}

	timer := t.mgr.CreateSlotTimer(id, delay, repeat, cb, args, t)
	if timer == nil {
		ELog.ErrorAf("[Timer] CreateSlotTimer Erorr id = %v,delay = %v", id, delay)
		return false
	}

	if delay == 0 {
		ELog.WarnA("[Timer] Delay = 0")
		timer.Call()
		return true
	}

	t.registerMap[id] = timer
	t.mgr.AddSlotTimer(timer)
	return true
}

package etimer

import (
	"runtime"
)

type TimerRegister struct {
	registerMap map[uint32]*Timer
	registerId  uint64
}

func CleanTimer(t *TimerRegister) {
	t.KillAllTimer()
}

func NewTimerRegister() *TimerRegister {
	tr := &TimerRegister{
		registerMap: make(map[uint32]*Timer),
		registerId:  0,
	}
	runtime.SetFinalizer(tr, CleanTimer)
	return tr
}

func (t *TimerRegister) AddOnceTimer(id uint32, delay uint64, desc string, cb FuncType, args ArgType, replace bool) {
	t.addTimer(id, delay, desc, false, replace, cb, args)
}

func (t *TimerRegister) AddRepeatTimer(id uint32, delay uint64, desc string, cb FuncType, args ArgType, replace bool) {
	t.addTimer(id, delay, desc, true, replace, cb, args)
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
	timer, ok := t.registerMap[info.registerEid]
	if ok {
		if timer.uid == info.uid && timer.registerUid == info.registerUid {
			delete(t.registerMap, timer.registerEid)
		}
	}
}

func (t *TimerRegister) addTimer(id uint32, delay uint64, desc string, repeat bool, replace bool, cb FuncType, args ArgType) bool {
	if delay == NovalidDelayMill {
		return false
	}

	exist := t.HasTimer(id)
	if exist && !replace {
		return true
	}

	if exist && replace {
		t.KillTimer(id)
	}

	t.registerId++
	timer := GTimerMgr.CreateSlotTimer(id, t.registerId, delay, desc, repeat, cb, args, t)
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
	GTimerMgr.AddSlotTimer(timer)
	return true
}

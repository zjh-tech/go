package etimer

import (
	"projects/go-engine/elog"
	"runtime"
)

type TimerRegister struct {
	register_map map[uint32]*Timer
	register_id  uint64
}

func CleanTimer(t *TimerRegister) {
	t.KillAllTimer()
}

func NewTimerRegister() *TimerRegister {
	tr := &TimerRegister{
		register_map: make(map[uint32]*Timer),
		register_id:  0,
	}
	runtime.SetFinalizer(tr, CleanTimer)
	return tr
}

func (t *TimerRegister) AddOnceTimer(id uint32, delay uint64, desc string, cb FuncType, args ArgType, replace bool) {
	t.add_timer(id, delay, desc, false, replace, cb, args)
}

func (t *TimerRegister) AddRepeatTimer(id uint32, delay uint64, desc string, cb FuncType, args ArgType, replace bool) {
	t.add_timer(id, delay, desc, true, replace, cb, args)
}

func (t *TimerRegister) HasTimer(id uint32) bool {
	_, ok := t.register_map[id]
	return ok
}

func (t *TimerRegister) GetRemainTime(id uint32) (bool, uint64) {
	timer, ok := t.register_map[id]
	if !ok {
		return false, uint64(0)
	}

	return true, timer.get_remain_time()
}

func (t *TimerRegister) KillTimer(id uint32) {
	timer, ok := t.register_map[id]
	if ok {
		timer.Kill()
		delete(t.register_map, id)
	}
}

func (t *TimerRegister) KillAllTimer() {
	for _, timer := range t.register_map {
		timer.Kill()
	}
	t.register_map = make(map[uint32]*Timer)
}

func (t *TimerRegister) RemoveTimer(info *Timer) {
	timer, ok := t.register_map[info.register_eid]
	if ok {
		if timer.uid == info.uid && timer.register_uid == info.register_uid {
			delete(t.register_map, timer.register_eid)
		}
	}
}

func (t *TimerRegister) add_timer(id uint32, delay uint64, desc string, repeat bool, replace bool, cb FuncType, args ArgType) bool {
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

	t.register_id++
	timer := GTimerMgr.CreateSlotTimer(id, t.register_id, delay, desc, repeat, cb, args, t)
	if timer == nil {
		elog.ErrorAf("[Timer] CreateSlotTimer Erorr id = %v,delay = %v", id, delay)
		return false
	}

	if delay == 0 {
		elog.WarnA("[Timer] Delay = 0")
		timer.Call()
		return true
	}

	t.register_map[id] = timer
	GTimerMgr.AddSlotTimer(timer)
	return true
}

package etimer

import (
	"reflect"
)

type Timer struct {
	uid      uint64
	eid      uint32
	delay    uint64
	repeat   bool
	rotation int64
	slot     uint64
	cb       FuncType
	args     ArgType
	state    TimerState
	register *TimerRegister
}

func newTimer(eid uint32, uid uint64, delay uint64, repeat bool, cb FuncType, args ArgType, register *TimerRegister) *Timer {
	timer := &Timer{
		eid:      eid,
		uid:      uid,
		delay:    delay,
		repeat:   repeat,
		cb:       cb,
		args:     args,
		state:    TimerInvalidState,
		register: register,
	}
	return timer
}

func (t *Timer) Kill() {
	t.state = TimerKilledState
	ELog.DebugAf("[Timer] id %v-%v Kill State", t.uid, t.eid)
}

func (t *Timer) Call() {
	defer func() {
		if err := recover(); err != nil {
			ELog.ErrorAf("[Timer] func%v args:%v call err: %v", reflect.TypeOf(t.cb).Name(), t.args, err)
		}
	}()

	t.cb(t.args...)
}

func (t *Timer) getRemainTime() uint64 {
	remainTime := uint64(0)
	if t.state != TimerRunningState {
		return remainTime
	}

	curSlot := GTimerMgr.GetCurSlot()
	if curSlot < t.slot {
		remainTime = uint64(t.rotation)*MaxSlotSize + t.slot - curSlot
	} else {
		remainTime = uint64(t.rotation)*MaxSlotSize + (MaxSlotSize - curSlot + t.slot)
	}

	return remainTime
}

package etimer

import (
	"reflect"
)

type Timer struct {
	uid         uint64
	registerEid uint32
	registerUid uint64
	delay       uint64
	desc        string
	repeat      bool
	rotation    int64
	slot        uint64
	cb          FuncType
	args        ArgType
	state       TimerState
	register    *TimerRegister
}

func new_timer(registerEid uint32, registerUid uint64, uid uint64, delay uint64, desc string, repeat bool, cb FuncType, args ArgType, register *TimerRegister) *Timer {
	timer := &Timer{
		registerEid: registerEid,
		registerUid: registerUid,
		uid:         uid,
		delay:       delay,
		desc:        desc,
		repeat:      repeat,
		cb:          cb,
		args:        args,
		state:       TimerInvalidState,
		register:    register,
	}
	return timer
}

func (t *Timer) Kill() {
	t.state = TimerKilledState
	ELog.DebugAf("[Timer] desc=%v id %v-%v-%v Kill State", t.desc, t.registerUid, t.uid, t.registerEid)
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

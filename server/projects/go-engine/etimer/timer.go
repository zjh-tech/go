package etimer

import (
	"projects/go-engine/elog"
	"reflect"
)

type TimerState int32

const (
	TimerInvalidState TimerState = iota
	TimerRunningState
	TimerKilledState
)

type Timer struct {
	uid          uint64
	register_eid uint32
	register_uid uint64
	delay        uint64
	desc         string
	repeat       bool
	rotation     int64
	slot         uint64
	cb           FuncType
	args         ArgType
	state        TimerState
	register     *TimerRegister
}

func new_timer(register_eid uint32, register_uid uint64, uid uint64, delay uint64, desc string, repeat bool, cb FuncType, args ArgType, register *TimerRegister) *Timer {
	timer := &Timer{
		register_eid: register_eid,
		register_uid: register_uid,
		uid:          uid,
		delay:        delay,
		desc:         desc,
		repeat:       repeat,
		cb:           cb,
		args:         args,
		state:        TimerInvalidState,
		register:     register,
	}
	return timer
}

func (t *Timer) Kill() {
	t.state = TimerKilledState
	elog.DebugAf("[Timer] desc=%v id %v-%v-%v Kill State", t.desc, t.register_uid, t.uid, t.register_eid)
}

func (t *Timer) Call() {
	defer func() {
		if err := recover(); err != nil {
			elog.ErrorAf("[Timer] func%v args:%v call err: %v", reflect.TypeOf(t.cb).Name(), t.args, err)
		}
	}()

	t.cb(t.args...)
}

func (t *Timer) get_remain_time() uint64 {
	remain_time := uint64(0)
	if t.state != TimerRunningState {
		return remain_time
	}

	cur_slot := GTimerMgr.GetCurSlot()
	if cur_slot < t.slot {
		remain_time = uint64(t.rotation)*MaxSlotSize + t.slot - cur_slot
	} else {
		remain_time = uint64(t.rotation)*MaxSlotSize + (MaxSlotSize - cur_slot + t.slot)
	}

	return remain_time
}

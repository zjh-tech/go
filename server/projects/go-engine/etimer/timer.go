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
	Id          uint32
	RegisterId  uint64
	TimeWheelId uint64
	Delay       uint64
	Desc        string
	Repeat      bool
	Rotation    int64
	Slot        uint64
	Cb          FuncType
	Args        ArgType
	State       TimerState
	Register    *TimerRegister
}

func NewTimer(id uint32, registerId uint64, timeWheelId uint64, delay uint64, desc string, repeat bool, cb FuncType, args ArgType, register *TimerRegister) *Timer {
	timer := &Timer{
		Id:          id,
		RegisterId:  registerId,
		TimeWheelId: timeWheelId,
		Delay:       delay,
		Desc:        desc,
		Repeat:      repeat,
		Cb:          cb,
		Args:        args,
		State:       TimerInvalidState,
		Register:    register,
	}
	return timer
}

func (t *Timer) Kill() {
	t.State = TimerKilledState
	elog.DebugAf("[Timer] desc=%v id %v-%v-%v Kill State", t.Desc, t.RegisterId, t.TimeWheelId, t.Id)
}

func (t *Timer) Call() {
	defer func() {
		if err := recover(); err != nil {
			elog.ErrorAf("[Timer] func%v args:%v call err: %v", reflect.TypeOf(t.Cb).Name(), t.Args, err)
		}
	}()

	t.Cb(t.Args...)
}

func (t *Timer) GetRemainTime() uint64 {
	remainTime := uint64(0)
	if t.State != TimerRunningState {
		return remainTime
	}

	curSlot := GTimerMgr.GetCurSlot()
	if curSlot < t.Slot {
		remainTime = uint64(t.Rotation)*MaxSlotSize + t.Slot - curSlot
	} else {
		remainTime = uint64(t.Rotation)*MaxSlotSize + (MaxSlotSize - curSlot + t.Slot)
	}

	return remainTime
}

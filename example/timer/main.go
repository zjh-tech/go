package main

import (
	"projects/frame"
	"projects/engine/elog"
	"projects/engine/etimer"
)

type TimerID int32

const (
	TEST1_TIMER_ID TimerID = 1
	TEST2_TIMER_ID TimerID = 2
)

type Test struct {
	timeRegister etimer.ITimerRegister
}

func NewTest() *Test {
	t := &Test{
		timeRegister: etimer.NewTimerRegister(),
	}
	return t
}

func (t *Test) TestFunc() {
	a := 1
	b := 2

	t.timeRegister.AddOnceTimer(uint32(TEST1_TIMER_ID), 1000*60, "Test1", func(v ...interface{}) {
		tempA := v[0].(int)
		tempB := v[1].(int)
		ELog.InfoAf("TEST1 Exec a=%v  b=%v", tempA, tempB)
	}, []interface{}{a, b}, true)

	t.timeRegister.AddRepeatTimer(uint32(TEST2_TIMER_ID), 1000*30, "Test2", func(v ...interface{}) {
		ELog.InfoAf("TestID2 Exec")
	}, []interface{}{}, true)
}

func main() {
	logger := elog.NewLogger("./log", 0)
	logger.Init()
	ELog.SetLogger(logger)

	test := NewTest()
	test.TestFunc()

	for {
		etimer.GTimerMgr.Update(frame.TIMER_LOOP_COUNT)
	}
}

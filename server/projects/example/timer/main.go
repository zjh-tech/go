package main

import (
	"projects/go-engine/elog"
	"projects/go-engine/etimer"
)

type TimerID int32

const (
	TestTestID TimerID = 1
)

type TestTimer struct {
	timeRegister etimer.ITimerRegister
}

func NewTestTimer() *TestTimer {
	t := &TestTimer{
		timeRegister: etimer.NewTimerRegister(),
	}
	return t
}

func (t *TestTimer) TestRepeatTimer() {
	a := 1
	b := 2

	//t.timeRegister.AddRepeatTimer(uint32(TestTestID), 1000*60, "TestA", func(v ...interface{}) {
	//	//tempA := v[0].(int)
	//	//tempB := v[1].(int)
	//	//elog.InfoAf("TestRepeatTimer a=%v  b=%v", tempA, tempB)
	//}, []interface{}{a, b}, true)

	t.timeRegister.AddRepeatTimer(uint32(TestTestID), 1000*90, "TestB", func(v ...interface{}) {
		//tempA := v[0].(int)
		//tempB := v[1].(int)
		//elog.InfoAf("TestRepeatTimer a=%v  b=%v", tempA, tempB)
	}, []interface{}{a, b}, true)

	//t.timeRegister.AddRepeatTimer(uint32(TestTestID), 1000*30, "TestC", func(v ...interface{}) {
	//	//tempA := v[0].(int)
	//	//tempB := v[1].(int)
	//	//elog.InfoAf("TestRepeatTimer a=%v  b=%v", tempA, tempB)
	//}, []interface{}{a, b}, true)
}

func main() {
	elog.InitLog("./log", 0, nil, nil)

	test := NewTestTimer()
	test.TestRepeatTimer()

	for {
		etimer.GTimerMgr.Update()
	}
}

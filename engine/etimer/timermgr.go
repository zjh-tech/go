package etimer

import (
	"container/list"
)

type TimerMgr struct {
	uid      uint64
	slotList [MaxSlotSize]*list.List
	curSlot  uint64
	lastTick int64
}

func NewTimerMgr() *TimerMgr {
	mgr := &TimerMgr{
		curSlot: 0,
		uid:     0,
	}

	for i := uint64(0); i < MaxSlotSize; i++ {
		mgr.slotList[i] = list.New()
	}

	mgr.lastTick = getMillSecond()
	return mgr
}

func (t *TimerMgr) Update(loopCount int) bool {
	curMillSecond := getMillSecond()
	if curMillSecond < t.lastTick {
		ELog.ErrorA("[Timer] Time Rollback")
		return false
	}

	busy := false
	delta := curMillSecond - t.lastTick
	if delta > int64(loopCount) {
		delta = int64(loopCount)
		ELog.WarnA("[Timer] Time Forward")
	}
	t.lastTick += delta

	for i := int64(0); i < delta; i++ {
		t.curSlot++
		t.curSlot = t.curSlot % MaxSlotSize

		slotList := t.slotList[t.curSlot]
		var next *list.Element
		repeat_list := list.New()
		for e := slotList.Front(); e != nil; {
			next = e.Next()
			timer := e.Value.(*Timer)
			if timer.state != TimerRunningState {
				t.ReleaseTimer(timer)
				slotList.Remove(e)
				e = next
				continue
			}

			timer.rotation--
			if timer.rotation < 0 {
				busy = true
				ELog.DebugAf("[Timer] Trigger  id %v-%v", timer.uid, timer.eid)
				slotList.Remove(e)
				timer.Call()
				if timer.repeat && timer.state == TimerRunningState {
					repeat_list.PushBack(timer) //先加入repeatList,防止此循环又被遍历到
				} else {
					t.ReleaseTimer(timer)
				}
			} else {
				ELog.DebugAf("[Timer]  id %v-%v-%v remain rotation = %v %v", timer.uid, timer.eid, timer.rotation+1, MaxSlotSize)
			}

			e = next
		}

		if repeat_list.Len() != 0 {
			for e := repeat_list.Front(); e != nil; e = e.Next() {
				timer := e.Value.(*Timer)
				//不考虑timer.Call花费的时间，不然会有逻辑顺序问题
				t.AddSlotTimer(timer)
			}
		}
	}
	return busy
}

func (t *TimerMgr) UnInit() {
	ELog.Info("[Timer] Stop")
}

func (t *TimerMgr) CreateSlotTimer(eid uint32, delay uint64, repeat bool, cb FuncType, args ArgType, r *TimerRegister) *Timer {
	t.uid++
	timer := newTimer(eid, t.uid, delay, repeat, cb, args, r)
	return timer
}

func (t *TimerMgr) AddSlotTimer(timer *Timer) {
	if timer == nil {
		return
	}

	timer.state = TimerRunningState
	timer.rotation = int64(timer.delay / MaxSlotSize)
	timer.slot = (t.curSlot + timer.delay%MaxSlotSize) % MaxSlotSize
	tempRotation := timer.rotation
	if timer.slot == t.curSlot && timer.rotation > 0 {
		timer.rotation--
	}
	t.slotList[timer.slot].PushBack(timer)
	ELog.DebugAf("[Timer] AddSlotTimer  id %v-%v-%v delay=%v,curslot=%v,slot=%v,rotation=%v", timer.uid, timer.eid, timer.delay, t.curSlot, timer.slot, tempRotation)
}

func (t *TimerMgr) ReleaseTimer(timer *Timer) {
	if timer != nil {
		if timer.state == TimerRunningState {
			ELog.DebugAf("[Timer] ReleaseTimer  id %v-%v Running State", timer.uid, timer.eid)
		} else if timer.state == TimerKilledState {
			ELog.DebugAf("[Timer] ReleaseTimer id %v-%v Killed State", timer.uid, timer.eid)
		} else {
			ELog.DebugAf("[Timer] ReleaseTimer id %v-%v Unknow State", timer.uid, timer.eid)
		}

		timer.cb = nil
		timer.args = nil

		if timer.register != nil {
			timer.register.RemoveTimer(timer)
			timer.register = nil
		}
	}
}

func (t *TimerMgr) GetCurSlot() uint64 {
	return t.curSlot
}

var GTimerMgr *TimerMgr

func init() {
	GTimerMgr = NewTimerMgr()
}

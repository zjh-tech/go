package etimer

import (
	"container/list"
	"time"
)

func get_mill_second() int64 {
	return time.Now().UnixNano() / 1e6
}

type TimerMgr struct {
	uid       uint64
	slot_list [MaxSlotSize]*list.List
	cur_slot  uint64
	last_tick int64
}

func new_timer_mgr() *TimerMgr {
	mgr := &TimerMgr{
		cur_slot: 0,
		uid:      0,
	}

	for i := uint64(0); i < MaxSlotSize; i++ {
		mgr.slot_list[i] = list.New()
	}

	mgr.last_tick = get_mill_second()
	return mgr
}

func (t *TimerMgr) Update(loop_count int) bool {
	cur_mill_second := get_mill_second()
	if cur_mill_second < t.last_tick {
		ELog.ErrorA("[Timer] Time Rollback")
		return false
	}

	busy := false
	delta := cur_mill_second - t.last_tick
	if delta > int64(loop_count) {
		delta = int64(loop_count)
		ELog.WarnA("[Timer] Time Forward")
	}
	t.last_tick += delta

	for i := int64(0); i < delta; i++ {
		t.cur_slot++
		t.cur_slot = t.cur_slot % MaxSlotSize

		slot_list := t.slot_list[t.cur_slot]
		var next *list.Element
		repeat_list := list.New()
		for e := slot_list.Front(); e != nil; {
			next = e.Next()
			timer := e.Value.(*Timer)
			if timer.state != TimerRunningState {
				t.ReleaseTimer(timer)
				slot_list.Remove(e)
				e = next
				continue
			}

			timer.rotation--
			if timer.rotation < 0 {
				busy = true
				ELog.DebugAf("[Timer] Trigger %v id %v-%v-%v", timer.desc, timer.uid, timer.register_eid, timer.register_uid)
				slot_list.Remove(e)
				timer.Call()
				if timer.repeat && timer.state == TimerRunningState {
					repeat_list.PushBack(timer) //先加入repeatList,防止此循环又被遍历到
				} else {
					t.ReleaseTimer(timer)
				}
			} else {
				ELog.DebugAf("[Timer] %v id %v-%v-%v remain rotation = %v %v", timer.desc, timer.uid, timer.register_eid, timer.register_uid, timer.rotation+1, MaxSlotSize)
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

func (t *TimerMgr) CreateSlotTimer(register_eid uint32, register_uid uint64, delay uint64, desc string, repeat bool, cb FuncType, args ArgType, r *TimerRegister) *Timer {
	t.uid++
	timer := new_timer(register_eid, register_uid, t.uid, delay, desc, repeat, cb, args, r)
	return timer
}

func (t *TimerMgr) AddSlotTimer(timer *Timer) {
	if timer == nil {
		return
	}

	timer.state = TimerRunningState
	timer.rotation = int64(timer.delay / MaxSlotSize)
	timer.slot = (t.cur_slot + timer.delay%MaxSlotSize) % MaxSlotSize
	tempRotation := timer.rotation
	if timer.slot == t.cur_slot && timer.rotation > 0 {
		timer.rotation--
	}
	t.slot_list[timer.slot].PushBack(timer)
	ELog.DebugAf("[Timer] AddSlotTimer %v id %v-%v-%v delay=%v,curslot=%v,slot=%v,rotation=%v", timer.desc, timer.uid, timer.register_eid, timer.register_uid, timer.delay, t.cur_slot, timer.slot, tempRotation)
}

func (t *TimerMgr) ReleaseTimer(timer *Timer) {
	if timer != nil {
		if timer.state == TimerRunningState {
			ELog.DebugAf("[Timer] ReleaseTimer %v id %v-%v-%v Running State", timer.desc, timer.uid, timer.register_eid, timer.register_uid)
		} else if timer.state == TimerKilledState {
			ELog.DebugAf("[Timer] ReleaseTimer %v id %v-%v-%v Killed State", timer.desc, timer.uid, timer.register_eid, timer.register_uid)
		} else {
			ELog.DebugAf("[Timer] ReleaseTimer %v id %v-%v-%v Unknow State", timer.desc, timer.uid, timer.register_eid, timer.register_uid)
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
	return t.cur_slot
}

var GTimerMgr *TimerMgr

func init() {
	GTimerMgr = new_timer_mgr()
}

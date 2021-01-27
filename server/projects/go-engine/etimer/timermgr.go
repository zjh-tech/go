package etimer

import (
	"container/list"
	"projects/go-engine/elog"
	"time"
)

const MaxSlotSize uint64 = 60000
const SlotInterValTime uint64 = 1
const NovalidDelayMill uint64 = 0xFFFFFFFFFFFFFFFF

func GetMillSecond() int64 {
	return time.Now().UnixNano() / 1e6
}

type TimerMgr struct {
	slotList    [MaxSlotSize]*list.List
	curSlot     uint64
	lastTick    int64
	timeWheelId uint64
}

func NewTimerMgr() *TimerMgr {
	mgr := &TimerMgr{
		curSlot:     0,
		timeWheelId: 0,
	}

	for i := uint64(0); i < MaxSlotSize; i++ {
		mgr.slotList[i] = list.New()
	}

	mgr.lastTick = GetMillSecond()
	return mgr
}

func (t *TimerMgr) Update(loop_count int) bool {
	curMillSecond := GetMillSecond()
	if curMillSecond < t.lastTick {
		elog.ErrorA("[Timer] Time Rollback")
		return false
	}

	busy := false
	delta := curMillSecond - t.lastTick
	if delta > int64(loop_count) {
		delta = int64(loop_count)
		elog.WarnA("[Timer] Time Forward")
	}
	t.lastTick += delta

	for i := int64(0); i < delta; i++ {
		t.curSlot++
		t.curSlot = t.curSlot % MaxSlotSize

		slotList := t.slotList[t.curSlot]
		var next *list.Element
		repeatList := list.New()
		for e := slotList.Front(); e != nil; {
			next = e.Next()
			timer := e.Value.(*Timer)
			if timer.State != TimerRunningState {
				t.ReleaseTimer(timer)
				slotList.Remove(e)
				e = next
				continue
			}

			timer.Rotation--
			if timer.Rotation < 0 {
				busy = true
				elog.DebugAf("[Timer] Trigger %v id %v-%v-%v", timer.Desc, timer.TimeWheelId, timer.Id, timer.RegisterId)
				slotList.Remove(e)
				timer.Call()
				if timer.Repeat && timer.State == TimerRunningState {
					repeatList.PushBack(timer) //先加入repeatList,防止此循环又被遍历到
				} else {
					t.ReleaseTimer(timer)
				}
			} else {
				elog.DebugAf("[Timer] %v id %v-%v-%v remain rotation = %v %v", timer.Desc, timer.TimeWheelId, timer.Id, timer.RegisterId, timer.Rotation+1, MaxSlotSize)
			}

			e = next
		}

		if repeatList.Len() != 0 {
			for e := repeatList.Front(); e != nil; e = e.Next() {
				timer := e.Value.(*Timer)
				//不考虑timer.Call花费的时间，不然会有逻辑顺序问题
				t.AddSlotTimer(timer)
			}
		}
	}
	return busy
}

func (t *TimerMgr) UnInit() {
	elog.Info("[Timer] Stop")
}

func (t *TimerMgr) CreateSlotTimer(id uint32, registerId uint64, delay uint64, desc string, repeat bool, cb FuncType, args ArgType, r *TimerRegister) *Timer {
	t.timeWheelId++
	timer := NewTimer(id, registerId, t.timeWheelId, delay, desc, repeat, cb, args, r)
	return timer
}

func (t *TimerMgr) AddSlotTimer(timer *Timer) {
	if timer == nil {
		return
	}

	timer.State = TimerRunningState
	timer.Rotation = int64(timer.Delay / MaxSlotSize)
	timer.Slot = (t.curSlot + timer.Delay%MaxSlotSize) % MaxSlotSize
	tempRotation := timer.Rotation
	if timer.Slot == t.curSlot && timer.Rotation > 0 {
		timer.Rotation--
	}
	t.slotList[timer.Slot].PushBack(timer)
	elog.DebugAf("[Timer] AddSlotTimer %v id %v-%v-%v delay=%v,curslot=%v,slot=%v,rotation=%v", timer.Desc, timer.TimeWheelId, timer.Id, timer.RegisterId, timer.Delay, t.curSlot, timer.Slot, tempRotation)
}

func (t *TimerMgr) ReleaseTimer(timer *Timer) {
	if timer != nil {
		if timer.State == TimerRunningState {
			elog.DebugAf("[Timer] ReleaseTimer %v id %v-%v-%v Running State", timer.Desc, timer.TimeWheelId, timer.Id, timer.RegisterId)
		} else if timer.State == TimerKilledState {
			elog.DebugAf("[Timer] ReleaseTimer %v id %v-%v-%v Killed State", timer.Desc, timer.TimeWheelId, timer.Id, timer.RegisterId)
		} else {
			elog.DebugAf("[Timer] ReleaseTimer %v id %v-%v-%v Unknow State", timer.Desc, timer.TimeWheelId, timer.Id, timer.RegisterId)
		}

		timer.Cb = nil
		timer.Args = nil

		if timer.Register != nil {
			timer.Register.RemoveTimer(timer)
			timer.Register = nil
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

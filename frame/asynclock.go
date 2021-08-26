package frame

import "github.com/zjh-tech/go-frame/base/util"

type AsyncLockItem struct {
	Lockkey       int
	StartLockTime int64
	EndLockTime   int64
}

type AsyncLockItemMgr struct {
	lockMap map[int]*AsyncLockItem
}

func (a *AsyncLockItemMgr) Lock(lockkey int, locktime int64) bool {
	now := util.GetSecond()
	if info, ok := a.lockMap[lockkey]; ok {
		if info.EndLockTime < now {
			return true
		}
		info.StartLockTime = now
		info.EndLockTime = now + locktime
		return false
	}

	item := &AsyncLockItem{
		Lockkey:       lockkey,
		StartLockTime: now,
		EndLockTime:   now + locktime,
	}
	a.lockMap[lockkey] = item
	return false
}

func (a *AsyncLockItemMgr) UnLock(lockkey int) {
	delete(a.lockMap, lockkey)
}

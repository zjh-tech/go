package etimer

import "fmt"

type TimerState int32

const (
	TimerInvalidState TimerState = iota
	TimerRunningState
	TimerKilledState
)

const MaxSlotSize uint64 = 60000
const SlotInterValTime uint64 = 1
const NovalidDelayMill uint64 = 0xFFFFFFFFFFFFFFFF

const TimerMajorVersion = 1
const TimerMinorVersion = 1

type TimerVersion struct {
}

func (t *TimerVersion) GetVersion() string {
	return fmt.Sprintf("Timer Version: %v.%v", TimerMajorVersion, TimerMinorVersion)
}

var GTimerVersion *TimerVersion

func init() {
	GTimerVersion = &TimerVersion{}
}

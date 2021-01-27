package frame

import (
	"projects/config"
	"projects/go-engine/elog"
)

type AsyncEvent struct {
	EventType uint32
	Datas     interface{}
}

func NewAsyncEvent(t uint32, datas interface{}) *AsyncEvent {
	return &AsyncEvent{
		EventType: t,
		Datas:     datas,
	}
}

const (
	ReloadConfigType uint32 = iota
)

type AsyncModule struct {
	asyncEvtQueue chan *AsyncEvent
}

func (c *AsyncModule) StartReloadAllConfig() {
	go func() {
		elog.InfoA("[Server] ReloadAllConfig Start")
		cfgMgr := &config.ConfigMgr{}
		if err := cfgMgr.LoadAllCfg(config.GConfigMgr.DirPath); err != nil {
			elog.ErrorAf("[Server] ReloadAllConfig Error=%v", err)
			return
		}

		evt := NewAsyncEvent(ReloadConfigType, cfgMgr)
		c.asyncEvtQueue <- evt
	}()
}

func (c *AsyncModule) Run() bool {
	select {
	case evt, _ := <-c.asyncEvtQueue:
		if evt.EventType == ReloadConfigType {
			cfgMgr := evt.Datas.(*config.ConfigMgr)
			config.GConfigMgr = cfgMgr
			elog.InfoA("[Server] ReloadAllConfig End")
		}
		return true
	default:
		return false
	}
}

var GAsyncModule *AsyncModule

func init() {
	GAsyncModule = &AsyncModule{
		asyncEvtQueue: make(chan *AsyncEvent, 100),
	}
}

package frame

import (
	"projects/config"
)

type AsyncEvent struct {
	event_type uint32
	datas      interface{}
}

func NewAsyncEvent(t uint32, datas interface{}) *AsyncEvent {
	return &AsyncEvent{
		event_type: t,
		datas:      datas,
	}
}

const (
	ReloadConfigType uint32 = iota
)

type AsyncModule struct {
	async_evt_queue chan *AsyncEvent
}

func new_async_module(evt_queue_size int) *AsyncModule {
	return &AsyncModule{
		async_evt_queue: make(chan *AsyncEvent, evt_queue_size),
	}
}

func (c *AsyncModule) Run() bool {
	select {
	case evt, _ := <-c.async_evt_queue:
		if evt.event_type == ReloadConfigType {
			cfgMgr := evt.datas.(*config.ConfigMgr)
			config.GConfigMgr = cfgMgr
			ELog.InfoA("[Server] ReloadAllConfig End")
		}
		return true
	default:
		return false
	}
}

func (c *AsyncModule) StartReloadAllConfig() {
	go func() {
		ELog.InfoA("[Server] ReloadAllConfig Start")
		cfg_mgr := &config.ConfigMgr{}
		if err := cfg_mgr.LoadAllCfg(config.GConfigMgr.DirPath); err != nil {
			ELog.ErrorAf("[Server] ReloadAllConfig Error=%v", err)
			return
		}

		evt := NewAsyncEvent(ReloadConfigType, cfg_mgr)
		c.async_evt_queue <- evt
	}()
}

var GAsyncModule *AsyncModule

func init() {
	GAsyncModule = new_async_module(1024)
}

package main

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
	ReloadRegistryCfgType uint32 = iota
)

type AsyncModule struct {
	async_evt_queue chan *AsyncEvent
}

func new_async_module(evt_queue_size int) *AsyncModule {
	return &AsyncModule{
		async_evt_queue: make(chan *AsyncEvent, evt_queue_size),
	}
}

func (c *AsyncModule) Run(loop_count int) bool {
	for i := 0; i < loop_count; i++ {
		select {
		case evt, _ := <-c.async_evt_queue:
			if evt.event_type == ReloadRegistryCfgType {
				cfg := evt.datas.(*RegistryCfg)
				GRegistryCfg = cfg
				ELog.InfoA("[Server] ReloadRegistryCfg End")
			}
			return true
		default:
			return false
		}
	}
	ELog.ErrorA("[AsyncModule] Run Error")
	return false
}

func (c *AsyncModule) ReloadRegistryCfg(path string) {
	go func() {
		ELog.InfoA("[Server] ReloadRegistryCfg Start")
		cfg, err := ReadRegistryCfg(path)
		if err != nil {
			ELog.ErrorAf("[ServiceDiscovery]  ReloadRegistryCfg Error=%v", err)
		}
		evt := NewAsyncEvent(ReloadRegistryCfgType, cfg)
		c.async_evt_queue <- evt
	}()
}

var GAsyncModule *AsyncModule

func init() {
	GAsyncModule = new_async_module(1024)
}

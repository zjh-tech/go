package main

import (
	"sync"

	"github.com/zjh-tech/go-frame/base/util"
	"github.com/zjh-tech/go-frame/engine/etimer"
)

const (
	RELOAD_REGISTRY_CFG_TIMER_ID uint32 = iota
	WARN_SERVICE_TIMER_ID
	PRINT_USE_SERVICE_TIMER_ID
	REBUILD_SERVICE_LIST_TIMER_ID
)

const (
	RELOAD_REGISTRY_CFG_TIMER_DELAY  uint64 = 1000 * 60 * 2
	PRINT_USE_SERVICE_TIMER_DELAY    uint64 = 1000 * 60
	WARN_SERVICE_TIMER_DELAY         uint64 = 1000 * 1
	REBUILD_SERVICE_LIST_TIMER_DELAY uint64 = 1000 * 30
)

const (
	SERVICE_WARN_TIME        int64 = 1000 * 10
	SERVICE_MAX_TIMEOUT_TIME int64 = 1000 * 60 * 2
)

type ServiceData struct {
	ServerId uint64
	Tick     int64
	WarnFlag bool
}

type ServiceDiscoveryServer struct {
	timeRegister etimer.ITimerRegister
	UseServices  map[uint64]*ServiceData
	RebuildFlag  bool
	Mutex        sync.Mutex
}

func NewServiceDiscoveryServer() *ServiceDiscoveryServer {
	return &ServiceDiscoveryServer{
		timeRegister: etimer.NewTimerRegister(),
		UseServices:  make(map[uint64]*ServiceData),
		RebuildFlag:  true,
	}
}

func (s *ServiceDiscoveryServer) Init(path string) {
	s.timeRegister.AddOnceTimer(REBUILD_SERVICE_LIST_TIMER_ID, REBUILD_SERVICE_LIST_TIMER_DELAY, "SDServer-RebuildServiceList", func(v ...interface{}) {
		GServiceDiscoveryServer.RebuildFlag = false
		ELog.Info("[ServiceDiscovery] Rebuild  ServiceList OK")
	}, []interface{}{path}, true)

	//ReadRegistryCfg Timer
	s.timeRegister.AddRepeatTimer(RELOAD_REGISTRY_CFG_TIMER_ID, RELOAD_REGISTRY_CFG_TIMER_DELAY, "SDServer-ReadRegistryCfg", func(v ...interface{}) {
		cfgPath := v[0].(string)
		GAsyncModule.ReloadRegistryCfg(cfgPath)
	}, []interface{}{path}, true)

	//Print Use Service Timer
	s.timeRegister.AddRepeatTimer(PRINT_USE_SERVICE_TIMER_ID, PRINT_USE_SERVICE_TIMER_DELAY, "SDServer-PrintUseService", func(v ...interface{}) {
		GServiceDiscoveryServer.Mutex.Lock()
		defer GServiceDiscoveryServer.Mutex.Unlock()
		for _, usedServiceInfo := range GServiceDiscoveryServer.UseServices {
			ELog.InfoAf("[ServiceDiscovery] UsedServices=%+v", usedServiceInfo)
		}
	}, []interface{}{path}, true)

	//Add or Del WarnService Timer
	s.timeRegister.AddRepeatTimer(WARN_SERVICE_TIMER_ID, WARN_SERVICE_TIMER_DELAY, "SDServer-AddWarnAndDelUseService",
		func(v ...interface{}) {
			if GServiceDiscoveryServer.RebuildFlag == true {
				//重建中,不检查
				return
			}

			GServiceDiscoveryServer.Mutex.Lock()
			defer GServiceDiscoveryServer.Mutex.Unlock()

			now := util.GetMillsecond()

			for serverId, service := range GServiceDiscoveryServer.UseServices {
				if service.WarnFlag == false && service.Tick+SERVICE_WARN_TIME < now {
					ELog.InfoAf("[ServiceDiscovery] Add Warn ServerID=%v", serverId)
					service.WarnFlag = true
				}

				if service.WarnFlag == true && service.Tick+SERVICE_MAX_TIMEOUT_TIME < now {
					delete(GServiceDiscoveryServer.UseServices, serverId)
					ELog.InfoAf("[ServiceDiscovery] Del UsedService ServerID=%v", serverId)
				}
			}

		}, []interface{}{}, true)
}

func (s *ServiceDiscoveryServer) AddUsedService(ServerId uint64) {
	service := &ServiceData{
		ServerId: ServerId,
		Tick:     util.GetMillsecond(),
		WarnFlag: false,
	}
	s.UseServices[ServerId] = service
	ELog.InfoAf("[ServiceDiscovery] AddUsedService ServerId=%v", ServerId)
}

func (s *ServiceDiscoveryServer) RemoveWarnService(ServerId uint64) {
	if service, ok := s.UseServices[ServerId]; ok {
		service.Tick = util.GetMillsecond()
		if service.WarnFlag == true {
			service.WarnFlag = false
			ELog.InfoAf("[ServiceDiscovery] Remove WarnService %v", ServerId)
		}
	}
}

var GServiceDiscoveryServer *ServiceDiscoveryServer

func init() {
	GServiceDiscoveryServer = NewServiceDiscoveryServer()
}

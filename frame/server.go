package frame

import (
	"math/rand"
	"os"
	"time"

	"github.com/google/gops/agent"
	"github.com/zjh-tech/go-frame/base/util"
	"github.com/zjh-tech/go-frame/engine/ecommon"
	"github.com/zjh-tech/go-frame/engine/edb"
	"github.com/zjh-tech/go-frame/engine/elog"
	"github.com/zjh-tech/go-frame/engine/enet"
	"github.com/zjh-tech/go-frame/engine/eredis"
	"github.com/zjh-tech/go-frame/engine/etimer"
)

type IServerFacade interface {
	Init() bool

	Run()

	Quit()

	UnInit()
}

type Server struct {
	configPath      string
	terminate       bool
	localServerId   uint64
	localServerType uint32
	localIp         string
	localToken      string
	logger          *elog.Logger
}

func (s *Server) IsQuit() bool {
	return s.terminate
}

func (s *Server) Quit() {
	ELog.Info("Server Quit")
	s.terminate = true
}

func (s *Server) GetLocalServerID() uint64 {
	return s.localServerId
}

func (s *Server) GetLocalToken() string {
	return s.localToken
}

func (s *Server) GetLocalServerType() uint32 {
	return s.localServerType
}

func (s *Server) GetLocalIp() string {
	return s.localIp
}

func (s *Server) GetLogger() *elog.Logger {
	return s.logger
}

func (s *Server) GetConfigPath() string {
	return s.configPath
}

func (s *Server) Init() bool {
	configPath := "."
	if len(os.Args) == 2 {
		configPath = os.Args[1]
	}

	s.configPath = configPath
	s.terminate = false

	serverCfgPath := s.GetConfigPath() + "/serverCfg.xml"
	if serverCfg, readErr := ReadServerCfg(serverCfgPath); readErr != nil {
		return false
	} else {
		GServerCfg = serverCfg
	}

	s.localServerId = GServerCfg.ServerId
	s.localServerType = GServerCfg.ServerType
	s.localToken = GServerCfg.Token
	s.localIp, _ = util.GetLocalIp()

	//Log
	s.logger = elog.NewLogger(GServerCfg.LogDir, GServerCfg.LogLevel)
	s.logger.Init()
	s.initModulesLog()
	ELog.Info("Server Log System Init Success")
	s.printModulesVersion()

	rand.Seed(time.Now().UnixNano())

	//Signal
	GSignalDealer.Init(s)
	ELog.Info("Server Signal System Init Success")

	if s.GetLocalServerType() == 0 {
		ELog.Error("Server ServerType = 0 Error")
		return false
	}

	if GServerCfg.ServerId == 0 {
		ELog.Error("Server ServerID = 0")
		return false
	}

	//Uid
	idMaker, idErr := NewIdMaker(int64(s.localServerId))
	if idErr != nil {
		ELog.Errorf("Server IdMaker Error=%v", idMaker)
		return false
	}
	GIdMaker = idMaker

	//Gops
	gopsErr := agent.Listen(agent.Options{
		Addr:            "",
		ConfigDir:       "",
		ShutdownCleanup: true,
	})

	if gopsErr != nil {
		ELog.Errorf("Server Gops Error=%v", gopsErr)
		return false
	}

	GServer = s
	enet.GSSSessionMgr.Init(s.GetLocalIp())

	return true
}

func (s *Server) initModulesLog() {
	//IOC依赖注入
	ELog.SetLogger(s.logger)
	etimer.ELog.SetLogger(s.logger)
	edb.ELog.SetLogger(s.logger)
	enet.ELog.SetLogger(s.logger)
	eredis.ELog.SetLogger(s.logger)
}

func (s *Server) printModulesVersion() {
	modules := []ecommon.IVersion{elog.GLogVersion, edb.GDBVersion, eredis.GRedisVersion, etimer.GTimerVersion}
	for _, module := range modules {
		ELog.Info(module.GetVersion())
	}
}

func (s *Server) UnInit() {
	enet.GNet.UnInit()
	etimer.GTimerMgr.UnInit()
	s.logger.UnInit()
}

var GServer *Server

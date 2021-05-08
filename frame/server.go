package frame

import (
	"math/rand"
	"projects/base/util"
	"projects/engine/ecommon"
	"projects/engine/edb"
	"projects/engine/elog"
	"projects/engine/enet"
	"projects/engine/eredis"
	"projects/engine/etimer"
	"time"

	"github.com/google/gops/agent"
)

type IServerFacade interface {
	Init() bool

	Run()

	Quit()

	UnInit()
}

type Server struct {
	terminate         bool
	local_server_id   uint64
	local_server_type uint32
	local_ip          string
	local_token       string
	logger            *elog.Logger
}

func (s *Server) IsQuit() bool {
	return s.terminate
}

func (s *Server) Quit() {
	ELog.Info("Server Quit")
	s.terminate = true
}

func (s *Server) GetLocalServerID() uint64 {
	return s.local_server_id
}

func (s *Server) GetLocalToken() string {
	return s.local_token
}

func (s *Server) GetLocalServerType() uint32 {
	return s.local_server_type
}

func (s *Server) GetLocalIp() string {
	return s.local_ip
}

func (s *Server) GetLogger() *elog.Logger {
	return s.logger
}

func (s *Server) Init() bool {
	s.terminate = false

	err := ReadServerCfg("./server_cfg.xml")
	if err != nil {
		return false
	}

	s.local_server_id = GServerCfg.ServerId
	s.local_server_type = GServerCfg.ServerType
	s.local_token = GServerCfg.Token
	s.local_ip, _ = util.GetLocalIp()

	//Log
	s.logger = elog.NewLogger(GServerCfg.LogDir, GServerCfg.LogLevel)
	s.logger.Init()
	s.init_modules_log()
	ELog.Info("Server Log System Init Success")
	s.print_modules_version()

	rand.Seed(time.Now().UnixNano())

	//Signal
	GSignalDealer.RegisterSigHandler()
	GSignalDealer.SetSignalQuitDealer(s)
	GSignalDealer.ListenSignal()
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
	idMaker, idErr := NewIdMaker(int64(s.local_server_id))
	if idErr != nil {
		ELog.Errorf("Server IdMaker Error=%v", idMaker)
		return false
	}
	GIdMaker = idMaker

	//Gops
	gops_err := agent.Listen(agent.Options{
		Addr:            "",
		ConfigDir:       "",
		ShutdownCleanup: true,
	})

	if gops_err != nil {
		ELog.Errorf("Server Gops Error=%v", gops_err)
		return false
	}

	GServer = s
	GSSSessionMgr.Init()

	return true
}

func (s *Server) init_modules_log() {
	//IOC依赖注入
	ELog.SetLogger(s.logger)
	etimer.ELog.SetLogger(s.logger)
	edb.ELog.SetLogger(s.logger)
	enet.ELog.SetLogger(s.logger)
	eredis.ELog.SetLogger(s.logger)
}

func (s *Server) print_modules_version() {
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

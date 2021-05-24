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
	config_path       string
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

func (s *Server) GetConfigPath() string {
	return s.config_path
}

func (s *Server) Init() bool {
	config_path := "."
	if len(os.Args) == 2 {
		config_path = os.Args[1]
	}

	s.config_path = config_path
	s.terminate = false

	server_cfg_path := s.GetConfigPath() + "/server_cfg.xml"
	if server_cfg, read_err := ReadServerCfg(server_cfg_path); read_err != nil {
		return false
	} else {
		GServerCfg = server_cfg
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

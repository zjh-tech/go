package frame

import (
	"math/rand"
	"projects/go-engine/elog"
	"projects/go-engine/enet"
	"projects/go-engine/etimer"
	"projects/util"
	"sync"
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
	wg                sync.WaitGroup
	local_server_id   uint64
	local_server_type uint32
	local_ip          string
	local_token       string
}

func (s *Server) IsQuit() bool {
	return s.terminate
}

func (s *Server) Quit() {
	elog.Info("Server Quit")
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
	elog.Init(GServerCfg.LogDir, GServerCfg.LogLevel, func(i ...interface{}) {
		s.wg.Add(1)
	})

	elog.Info("Server Log System Init Success")
	rand.Seed(time.Now().UnixNano())

	//Signal
	GSignalDealer.RegisterSigHandler()
	GSignalDealer.SetSignalQuitDealer(s)
	GSignalDealer.ListenSignal()
	elog.Info("Server Signal System Init Success")

	if s.GetLocalServerType() == DEFAULT_SERVER_TYPE {
		elog.Error("Server ServerType = 0 Error")
		return false
	}

	if GServerCfg.ServerId == 0 {
		elog.Error("Server ServerID = 0")
		return false
	}

	//Uid
	idMaker, idErr := NewIdMaker(int64(s.local_server_id))
	if idErr != nil {
		elog.Errorf("Server IdMaker Error=%v", idMaker)
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
		elog.Errorf("Server Gops Error=%v", gopsErr)
		return false
	}

	GServer = s
	GSSServerSessionMgr.Init()

	return true
}

func (s *Server) UnInit() {
	enet.GNet.UnInit()
	etimer.GTimerMgr.UnInit()
	elog.UnInit(func(i ...interface{}) {
		s.wg.Done()
	})
	s.wg.Wait()
}

var GServer *Server

package frame

import (
	"math/rand"
	"os"
	"runtime"
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

type IServer interface {
	Init() bool

	Run()

	Quit()

	UnInit()
}

type Server struct {
	configPath string
	terminate  bool
	srvCfg     *ServerCfg
	logger     *elog.Logger
	ip         string
	state      uint32
}

func (s *Server) IsQuit() bool {
	return s.terminate
}

func (s *Server) Quit() {
	ELog.Info("Server Quit")
	s.terminate = true
}

func (s *Server) GetServerId() uint64 {
	return s.srvCfg.ServerInfo.ServerId
}

func (s *Server) GetToken() string {
	return s.srvCfg.ServerInfo.Token
}

func (s *Server) GetServerType() uint32 {
	return s.srvCfg.ServerInfo.ServerType
}

func (s *Server) GetIp() string {
	return s.ip
}

func (s *Server) GetSrvCfg() *ServerCfg {
	return s.srvCfg
}

func (s *Server) GetLogger() *elog.Logger {
	return s.logger
}

func (s *Server) GetConfigPath() string {
	return s.configPath
}

func (s *Server) GetState() uint32 {
	return s.state
}

func (s *Server) SetState(state uint32) {
	s.state = state
}

func (s *Server) Init() bool {
	configPath := "."
	if len(os.Args) == 2 {
		configPath = os.Args[1]
	}

	s.configPath = configPath
	s.terminate = false

	serverCfgPath := s.GetConfigPath() + "/config/server_cfg.yaml"
	if serverCfg, readErr := ReadServerCfg(serverCfgPath); readErr != nil {
		return false
	} else {
		s.srvCfg = serverCfg
	}

	s.ip, _ = util.GetLocalIp()

	rand.Seed(time.Now().UTC().UnixNano())
	runtime.GOMAXPROCS(runtime.NumCPU())

	//Log
	s.logger = elog.NewLogger(s.srvCfg.LogInfo.Path, s.srvCfg.LogInfo.Level)
	s.logger.Init()
	s.initModulesLog()
	ELog.Info("Server Log System Init Success")
	s.printModulesVersion()

	if s.GetServerType() == 0 {
		ELog.Error("Server ServerType = 0 Error")
		return false
	}

	if s.GetServerId() == 0 {
		ELog.Error("Server ServerID = 0")
		return false
	}

	//Redis
	if s.srvCfg.IsOpenRedis() {
		redisAddrs := s.srvCfg.GetRedisAddrs()
		if len(redisAddrs) == 0 {
			ELog.Error("Redis redisAddrs=%v Is Empty")
			return false
		}

		if s.srvCfg.IsOpenRedisCluster() {
			if redisClient, err := eredis.ConnectRedisCluster(redisAddrs, s.srvCfg.RedisInfo.Password); err != nil {
				ELog.Errorf("Redis redisAddrs=%v Connect Cluster Error", redisAddrs)
				return false
			} else {
				eredis.GRedisCmd = redisClient
				ELog.Infof("Redis redisAddrs=%v Connect Cluster Success", redisAddrs)
			}
		} else {
			if redisClient, err := eredis.ConnectRedis(redisAddrs[0], s.srvCfg.RedisInfo.Password); err != nil {
				ELog.Errorf("Redis redisAddr=%v Connect Error", redisAddrs[0])
				return false
			} else {
				eredis.GRedisCmd = redisClient
				ELog.Infof("Redis redisAddr=%v Connect Success", redisAddrs[0])
			}
		}
	}

	//Db
	if s.srvCfg.IsOpenDB() {
		if s.srvCfg.DBInfo != nil && s.srvCfg.DBInfo.ConnMaxCount != 0 && s.srvCfg.DBInfo.TableMaxCount != 0 {
			if err := edb.GDBModule.Init(s.srvCfg.DBInfo.ConnMaxCount, s.srvCfg.DBInfo.TableMaxCount, s.srvCfg.DBInfo.DBInfoList); err != nil {
				ELog.Error(err)
				return false
			}
		}
	}

	//Signal
	GSignalDealer.Init(s)
	ELog.Info("Server Signal System Init Success")

	//Uid
	idMaker, idErr := NewIdMaker(int64(s.GetServerId()), true)
	if idErr != nil {
		ELog.Errorf("Server IdMaker Error=%v", idErr)
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

	//Tcp Listen
	GServer = s
	enet.GSSSessionMgr.Init(s.GetToken())
	if len(s.srvCfg.ServerInfo.Outer) != 0 {
		if !enet.GSSSessionMgr.SSServerListen(s.srvCfg.ServerInfo.Outer) {
			ELog.ErrorA("Server Listen Outer=%v", s.srvCfg.ServerInfo.Outer)
			return false
		}
	} else if len(s.srvCfg.ServerInfo.Inter) != 0 {
		if !enet.GSSSessionMgr.SSServerListen(s.srvCfg.ServerInfo.Inter) {
			ELog.ErrorA("Server Listen Inter=%v", s.srvCfg.ServerInfo.Inter)
			return false
		}
	}

	//http
	if s.srvCfg.IsOpenHttp() {
		if len(s.srvCfg.HttpInfo.Cert) != 0 && len(s.srvCfg.HttpInfo.Key) != 0 {
			GSGLBClient.Init(s.srvCfg.HttpInfo.Url, true)
		} else {
			GSGLBClient.Init(s.srvCfg.HttpInfo.Url, false)
		}

		go func() {
			serviceRegisterTimer := time.NewTicker(30 * time.Second)
			defer serviceRegisterTimer.Stop()
			for {
				select {
				case <-serviceRegisterTimer.C:
					{
						spec := &ServiceRegisterReq{}
						spec.ServiceSpec = &ServiceSpec{}
						spec.ServiceSpec.ServiceID = s.GetServerId()
						spec.ServiceSpec.ServiceType = s.GetServerType()
						spec.ServiceSpec.Token = s.GetToken()
						spec.ServiceSpec.InterAddr = s.GetSrvCfg().ServerInfo.Inter
						spec.ServiceSpec.OuterAddr = s.GetSrvCfg().ServerInfo.Outer
						spec.ServiceSpec.State = s.GetState()
						GSGLBClient.SendServiceRegisterReq(spec)
					}
				}
			}
		}()
	}

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

package frame

import (
	"io/ioutil"

	"github.com/go-yaml/yaml"
	"github.com/zjh-tech/go-frame/engine/edb"
)

type ServerSpec struct {
	ServerId   uint64
	ServerType uint32
	Token      string
	Inter      string
	Outer      string
}

type LogSpec struct {
	Path  string
	Level int
}

type RedisConnSpec struct {
	Addr string
}

type RedisSpec struct {
	OpenFlag int              `yaml:"openflag"`
	Cluster  int              `yaml:"cluster"`
	Password string           `yaml:"password"`
	AddrList []*RedisConnSpec `yaml:"addrlist"`
}

type DBSpec struct {
	OpenFlag      int               `yaml:"openflag"`
	ConnMaxCount  uint64            `yaml:"connmaxcount"`
	TableMaxCount uint64            `yaml:"tablemaxcount"`
	DBInfoList    []*edb.DBConnSpec `yaml:"dblist"`
}

type HttpSpec struct {
	OpenFlag int    `yaml:"openflag"`
	Url      string `yaml:"url"`
	Cert     string `yaml:"cert"`
	Key      string `yaml:"key"`
}

type ServerCfg struct {
	ServerInfo ServerSpec `yaml:"server"`
	LogInfo    LogSpec    `yaml:"log"`
	RedisInfo  *RedisSpec `yaml:"redis"`
	DBInfo     *DBSpec    `yaml:"db"`
	HttpInfo   *HttpSpec  `yaml:"http"`
}

func NewServerCfg() *ServerCfg {
	return &ServerCfg{}
}

func (s *ServerCfg) GetRedisAddrs() []string {
	addrs := make([]string, 0)
	for _, info := range s.RedisInfo.AddrList {
		addrs = append(addrs, info.Addr)
	}
	return addrs
}

func (s *ServerCfg) IsOpenRedis() bool {
	return s.RedisInfo.OpenFlag != 0
}

func (s *ServerCfg) IsOpenRedisCluster() bool {
	return s.RedisInfo.Cluster != 0
}

func (s *ServerCfg) IsOpenDB() bool {
	return s.DBInfo.OpenFlag != 0
}

func (s *ServerCfg) IsOpenHttp() bool {
	return s.HttpInfo.OpenFlag != 0
}

func ReadServerCfg(path string) (*ServerCfg, error) {
	content, readErr := ioutil.ReadFile(path)
	if readErr != nil {
		return nil, readErr
	}
	serverCfg := NewServerCfg()
	if err := yaml.Unmarshal(content, &serverCfg); err != nil {
		return nil, err
	}

	return serverCfg, nil
}

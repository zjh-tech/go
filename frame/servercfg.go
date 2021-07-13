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
	Password string           `yaml:"password"`
	AddrList []*RedisConnSpec `yaml:"addrlist"`
}

type DBSpec struct {
	ConnMaxCount  uint64            `yaml:"connmaxcount"`
	TableMaxCount uint64            `yaml:"tablemaxcount"`
	DBInfoList    []*edb.DBConnSpec `yaml:"dblist"`
}

type HttpSpec struct {
	Url  string
	Cert string
	Key  string
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

func ReadServerCfg(path string) (*ServerCfg, error) {
	content, _ := ioutil.ReadFile(path)
	serverCfg := NewServerCfg()
	if err := yaml.Unmarshal(content, &serverCfg); err != nil {
		return nil, err
	}

	return serverCfg, nil
}

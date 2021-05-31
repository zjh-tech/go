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

type RedisSpec struct {
	Name     string
	Host     string
	Port     int
	Password string
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

func ReadServerCfg(path string) (*ServerCfg, error) {
	content, _ := ioutil.ReadFile(path)
	serverCfg := NewServerCfg()
	if err := yaml.Unmarshal(content, &serverCfg); err != nil {
		return nil, err
	}

	return serverCfg, nil
}

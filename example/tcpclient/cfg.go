package main

import (
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

type LogSpec struct {
	Path  string
	Level int
}

type TcpSpec struct {
	Addr string `yaml:"addr"`
}

type Cfg struct {
	LogInfo     LogSpec `yaml:"log"`
	TcpInfo     TcpSpec `yaml:"tcp"`
	ClientCount int     `yaml:"clientcount"`
	LoopCount   int     `yaml:"loopcount"`
}

func NewCfg() *Cfg {
	return &Cfg{}
}

func ReadCfg(path string) (*Cfg, error) {
	content, readErr := ioutil.ReadFile(path)
	if readErr != nil {
		return nil, readErr
	}
	Cfg := NewCfg()
	if err := yaml.Unmarshal(content, &Cfg); err != nil {
		return nil, err
	}

	return Cfg, nil
}

var GCfg *Cfg

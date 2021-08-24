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
	Addr         string `yaml:"addr"`
	MaxConnCount int    `yaml:"maxconncount"`
	OpenoverLoad int    `yaml:"openoverload"`
	Intervaltime int64  `yaml:"intervaltime"`
	Limit        int64  `yaml:"limit"`
}

type Cfg struct {
	LogInfo LogSpec `yaml:"log"`
	TcpInfo TcpSpec `yaml:"tcp"`
}

func NewCfg() *Cfg {
	return &Cfg{}
}

func ReadCfg(path string) (*Cfg, error) {
	content, readErr := ioutil.ReadFile(path)
	if readErr != nil {
		return nil, readErr
	}
	SdkCfg := NewCfg()
	if err := yaml.Unmarshal(content, &SdkCfg); err != nil {
		return nil, err
	}

	return SdkCfg, nil
}

var GCfg *Cfg

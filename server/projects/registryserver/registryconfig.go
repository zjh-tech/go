package main

import (
	"errors"
	"fmt"
	"projects/thirds/etree"
	"projects/util"
	"strings"

	"projects/go-engine/elog"
)

type S2SAttr struct {
	ServerID      uint64
	ServerType    uint32
	ServerTypeStr string
	Inter         string
	Outer         string
	Token         string
}

func NewS2SAttr() *S2SAttr {
	return &S2SAttr{
		Inter: "",
		Outer: "",
	}
}

type C2SAttr struct {
	ServerID      uint64
	ServerType    uint32
	ServerTypeStr string
	//Tcp
	Inter    string
	Outer    string
	MaxCount int
	//Https
	C2SHttpsUrl  string
	C2SHttpsCert string
	C2SHttpsKey  string
}

func NewC2SAttr() *C2SAttr {
	return &C2SAttr{
		Inter: "",
		Outer: "",
	}
}

type RegistryCfg struct {
	ServerTypeMap map[string]uint32
	ConntionMap   map[uint32][]uint32
	S2SAttrMap    map[uint64]*S2SAttr
	C2SAttrMap    map[uint64]*C2SAttr
}

func NewRegistryCfg() *RegistryCfg {
	return &RegistryCfg{
		ServerTypeMap: make(map[string]uint32),
		ConntionMap:   make(map[uint32][]uint32),
		S2SAttrMap:    make(map[uint64]*S2SAttr),
		C2SAttrMap:    make(map[uint64]*C2SAttr),
	}
}

var GRegistryCfg *RegistryCfg

func ReadRegistryCfg(path string) (*RegistryCfg, error) {
	doc := etree.NewDocument()
	err := doc.ReadFromFile(path)
	if err != nil {
		return nil, err
	}

	cfgRoot := doc.SelectElement("config")
	if cfgRoot == nil {
		return nil, errors.New("service_registry Xml Config Error")
	}

	cfg := NewRegistryCfg()
	//ServerType
	servertypeElem := cfgRoot.SelectElement("servertype")
	if servertypeElem != nil {
		for _, serverElem := range servertypeElem.FindElements("server") {
			var serverType string
			var serverEnum uint32
			for _, attr := range serverElem.Attr {
				if attr.Key == "type" {
					serverType = attr.Value
				} else if attr.Key == "enum" {
					serverEnum, _ = util.Str2Uint32(attr.Value)
				} else {
					errStr := fmt.Sprintf("service_registry Xml %v Error", attr.Value)
					return nil, errors.New(errStr)
				}
			}
			cfg.ServerTypeMap[serverType] = serverEnum
		}
	}

	//Connection
	connectionElem := cfgRoot.SelectElement("connection")
	if connectionElem != nil {
		for _, serverElem := range connectionElem.FindElements("server") {
			var start string
			var end string
			for _, attr := range serverElem.Attr {
				if attr.Key == "start" {
					start = attr.Value
				} else if attr.Key == "end" {
					end = attr.Value
				} else {
					errStr := fmt.Sprintf("service_registry Xml  %v Error", attr.Value)
					return nil, errors.New(errStr)
				}
			}

			endSlice := strings.Split(end, ",")
			endSliceLen := len(endSlice)
			if endSliceLen != 0 {
				startServerType, StartOk := cfg.ServerTypeMap[start]
				if !StartOk {
					errorStr := fmt.Sprintf("ServerType(%v) Not Find In RegistryConfig", start)
					return nil, errors.New(errorStr)
				}

				endArray := make([]uint32, 0)
				for j := 0; j < endSliceLen; j++ {
					endSliceStr := endSlice[j]
					if endServerType, endOk := cfg.ServerTypeMap[endSliceStr]; !endOk {
						errorStr := fmt.Sprintf("ServerType(%v) Not Find In RegistryConfig", endSliceStr)
						return nil, errors.New(errorStr)
					} else {
						endArray = append(endArray, endServerType)
					}
				}

				cfg.ConntionMap[startServerType] = endArray
			}
		}
	}

	//s2slist
	serverlistElem := cfgRoot.SelectElement("s2slist")
	if serverlistElem != nil {
		for _, serverElem := range serverlistElem.FindElements("server") {
			S2SAttr := NewS2SAttr()
			for _, attr := range serverElem.Attr {
				if attr.Key == "id" {
					S2SAttr.ServerID, _ = util.Str2Uint64(attr.Value)
				} else if attr.Key == "type" {
					S2SAttr.ServerTypeStr = attr.Value
					if serverType, ok := cfg.ServerTypeMap[attr.Value]; ok {
						S2SAttr.ServerType = serverType
					} else {
						errorStr := fmt.Sprintf("ServerType(%v) s2slist Not Find In RegistryConfig", attr.Value)
						return nil, errors.New(errorStr)
					}
				} else if attr.Key == "inter" {
					S2SAttr.Inter = attr.Value
				} else if attr.Key == "outer" {
					S2SAttr.Outer = attr.Value
				} else if attr.Key == "token" {
					S2SAttr.Token = attr.Value
				} else {
					errStr := fmt.Sprintf("service_registry s2slist Xml %v Error", attr.Value)
					return nil, errors.New(errStr)
				}
			}
			cfg.S2SAttrMap[S2SAttr.ServerID] = S2SAttr
		}
	}

	//c2slist
	c2slistElem := cfgRoot.SelectElement("c2slist")
	if c2slistElem != nil {
		for _, serverElem := range c2slistElem.FindElements("server") {
			clientAttr := NewC2SAttr()
			for _, attr := range serverElem.Attr {
				if attr.Key == "id" {
					clientAttr.ServerID, _ = util.Str2Uint64(attr.Value)
				} else if attr.Key == "type" {
					clientAttr.ServerTypeStr = attr.Value
					if serverType, ok := cfg.ServerTypeMap[attr.Value]; ok {
						clientAttr.ServerType = serverType
					} else {
						errorStr := fmt.Sprintf("ServerType(%v) c2slist Not Find In RegistryConfig", attr.Value)
						return nil, errors.New(errorStr)
					}
				} else if attr.Key == "inter" {
					clientAttr.Inter = attr.Value
				} else if attr.Key == "outer" {
					clientAttr.Outer = attr.Value
				} else if attr.Key == "maxcount" {
					clientAttr.MaxCount, _ = util.Str2Int(attr.Value)
				} else if attr.Key == "https_url" {
					clientAttr.C2SHttpsUrl = attr.Value
				} else if attr.Key == "https_cert" {
					clientAttr.C2SHttpsCert = attr.Value
				} else if attr.Key == "https_key" {
					clientAttr.C2SHttpsKey = attr.Value
				} else {
					errStr := fmt.Sprintf("service_registry c2slist Xml %v Error", attr.Value)
					return nil, errors.New(errStr)
				}
			}
			cfg.C2SAttrMap[clientAttr.ServerID] = clientAttr
		}
	}

	elog.Infof("ServerTypeMap=%+v", cfg.ServerTypeMap)
	elog.Infof("ConntionMap=%+v", cfg.ConntionMap)
	for _, attr := range cfg.S2SAttrMap {
		elog.Infof("S2SAttrMap=%+v", attr)
	}

	return cfg, nil
}

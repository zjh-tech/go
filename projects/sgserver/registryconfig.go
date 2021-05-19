package main

import (
	"errors"
	"fmt"
	"strings"

	"github.com/beevik/etree"
	"github.com/zjh-tech/go-frame/base/convert"
)

type ConnAttr struct {
	ServerID      uint64
	ServerType    uint32
	ServerTypeStr string

	//S2S TCP
	S2S_TCP_Inter string
	S2S_TCP_Outer string
	Token         string

	//S2S HTTP
	S2S_Http_SUrl  string
	S2S_Http_CUrl1 string
	S2S_Http_CUrl2 string

	//C2S TCP
	C2S_TCP_Inter string
	C2S_TCP_Outer string

	//C2S Https
	C2SHttpsUrl  string
	C2SHttpsCert string
	C2SHttpsKey  string

	//SDK TCP
	SDK_TCP_Inter string
	SDK_TCP_Outer string
	//SDK Https
	SDKHttpsUrl  string
	SDKHttpsCert string
	SDKHttpsKey  string
}

func NewConnAttr() *ConnAttr {
	return &ConnAttr{
		ServerID:       0,
		ServerType:     0,
		ServerTypeStr:  "",
		S2S_TCP_Inter:  "",
		S2S_TCP_Outer:  "",
		Token:          "",
		S2S_Http_SUrl:  "",
		S2S_Http_CUrl1: "",
		S2S_Http_CUrl2: "",
		C2S_TCP_Inter:  "",
		C2S_TCP_Outer:  "",
		C2SHttpsUrl:    "",
		C2SHttpsCert:   "",
		C2SHttpsKey:    "",
		SDK_TCP_Inter:  "",
		SDK_TCP_Outer:  "",
		SDKHttpsUrl:    "",
		SDKHttpsCert:   "",
		SDKHttpsKey:    "",
	}
}

type RegistryCfg struct {
	ServerTypeMap map[string]uint32
	ConntionMap   map[uint32][]uint32
	AttrMap       map[uint64]*ConnAttr
}

func NewRegistryCfg() *RegistryCfg {
	return &RegistryCfg{
		ServerTypeMap: make(map[string]uint32),
		ConntionMap:   make(map[uint32][]uint32),
		AttrMap:       make(map[uint64]*ConnAttr),
	}
}

var GRegistryCfg *RegistryCfg

func ReadRegistryCfg(path string) (*RegistryCfg, error) {
	doc := etree.NewDocument()
	err := doc.ReadFromFile(path)
	if err != nil {
		return nil, err
	}

	cfg_root := doc.SelectElement("config")
	if cfg_root == nil {
		return nil, errors.New("service_registry Xml Config Error")
	}

	cfg := NewRegistryCfg()
	//ServerType
	servertype_elem := cfg_root.SelectElement("servertype")
	if servertype_elem != nil {
		for _, serverElem := range servertype_elem.FindElements("server") {
			var serverType string
			var serverEnum uint32
			for _, attr := range serverElem.Attr {
				if attr.Key == "type" {
					serverType = attr.Value
				} else if attr.Key == "enum" {
					serverEnum, _ = convert.Str2Uint32(attr.Value)
				} else {
					errStr := fmt.Sprintf("service_registry Xml %v Error", attr.Value)
					return nil, errors.New(errStr)
				}
			}
			cfg.ServerTypeMap[serverType] = serverEnum
		}
	}

	//Connection
	connection_elem := cfg_root.SelectElement("connection")
	if connection_elem != nil {
		for _, serverElem := range connection_elem.FindElements("server") {
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

	list_elem := cfg_root.SelectElement("list")
	if list_elem != nil {
		for _, serverElem := range list_elem.FindElements("server") {
			conn_attr := NewConnAttr()
			for _, attr := range serverElem.Attr {
				if attr.Key == "id" {
					conn_attr.ServerID, _ = convert.Str2Uint64(attr.Value)
				} else if attr.Key == "type" {
					conn_attr.ServerTypeStr = attr.Value
					if serverType, ok := cfg.ServerTypeMap[attr.Value]; ok {
						conn_attr.ServerType = serverType
					} else {
						error_str := fmt.Sprintf("ServerType(%v) list Not Find In RegistryConfig", attr.Value)
						return nil, errors.New(error_str)
					}
				} else if attr.Key == "c2s_http_url" {
					conn_attr.C2SHttpsUrl = attr.Value
				} else if attr.Key == "c2s_http_cert" {
					conn_attr.C2SHttpsCert = attr.Value
				} else if attr.Key == "c2s_http_key" {
					conn_attr.C2SHttpsKey = attr.Value
				} else if attr.Key == "c2s_tcp_inter" {
					conn_attr.C2S_TCP_Inter = attr.Value
				} else if attr.Key == "c2s_tcp_outer" {
					conn_attr.C2S_TCP_Outer = attr.Value
				} else if attr.Key == "sdk_http_url" {
					conn_attr.SDKHttpsUrl = attr.Value
				} else if attr.Key == "sdk_http_cert" {
					conn_attr.SDKHttpsCert = attr.Value
				} else if attr.Key == "sdk_http_key" {
					conn_attr.SDKHttpsKey = attr.Value
				} else if attr.Key == "sdk_tcp_inter" {
					conn_attr.SDK_TCP_Inter = attr.Value
				} else if attr.Key == "sdk_tcp_outer" {
					conn_attr.SDK_TCP_Outer = attr.Value
				} else if attr.Key == "s2s_tcp_inter" {
					conn_attr.S2S_TCP_Inter = attr.Value
				} else if attr.Key == "s2s_tcp_outer" {
					conn_attr.S2S_TCP_Outer = attr.Value
				} else if attr.Key == "s2s_http_surl" {
					conn_attr.S2S_Http_SUrl = attr.Value
				} else if attr.Key == "s2s_http_curl1" {
					conn_attr.S2S_Http_CUrl1 = attr.Value
				} else if attr.Key == "s2s_http_curl2" {
					conn_attr.S2S_Http_CUrl2 = attr.Value
				} else if attr.Key == "token" {
					conn_attr.Token = attr.Value
				} else if attr.Key == "necessary" {

				} else {
					errStr := fmt.Sprintf("service_registry list Xml %v Error", attr.Value)
					return nil, errors.New(errStr)
				}
			}
			cfg.AttrMap[conn_attr.ServerID] = conn_attr
		}
	}

	ELog.Infof("ServerTypeMap=%+v", cfg.ServerTypeMap)
	ELog.Infof("ConntionMap=%+v", cfg.ConntionMap)
	for _, attr := range cfg.AttrMap {
		ELog.Infof("AttrMap=%+v", attr)
	}

	return cfg, nil
}

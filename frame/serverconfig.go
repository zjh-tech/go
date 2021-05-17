package frame

import (
	"errors"

	"github.com/zjh-tech/go-frame/base/convert"
	"github.com/zjh-tech/go-frame/base/etree"
)

type ServerCfg struct {
	ServiceName string
	ServerType  uint32
	ServerId    uint64
	Token       string //ServiceDiscovery的Token
	//Log
	LogDir   string
	LogLevel int

	//SD Tcp
	SDClientAddr string
	SDServerAddr string

	//SD Http
	SDClientUrl string
	SDServerUrl string

	//S-S  HttpServer
	S2SHttpServerUrl string

	//S-S  HttpClient
	S2SHttpClientUrl1 string
	S2SHttpClientUrl2 string

	//C-S  TCP
	C2SInterListen string
	C2SOuterListen string

	//C-S  Https
	C2SHttpsUrl  string
	C2SHttpsCert string
	C2SHttpsKey  string

	//SDK  TCP
	SDK_TCP_INTER string
	SDK_TCP_OUTER string

	//SDK  HTTPS
	SDKHttpsUrl  string
	SDKHttpsCert string
	SDKHttpsKey  string
}

func NewServerCfg() *ServerCfg {
	return &ServerCfg{}
}

var GServerCfg *ServerCfg

func ReadServerCfg(path string) (*ServerCfg, error) {
	doc := etree.NewDocument()
	err := doc.ReadFromFile(path)
	if err != nil {
		return nil, err
	}

	cfg_root := doc.SelectElement("config")
	if cfg_root == nil {
		return nil, errors.New("server_cfg Xml Config Error")
	}

	cfg := NewServerCfg()

	servicename_elem := cfg_root.FindElement("servicename")
	if servicename_elem != nil {
		cfg.ServiceName = servicename_elem.Text()
	}

	servertype_elem := cfg_root.FindElement("servertype")
	if servertype_elem != nil {
		cfg.ServerType, _ = convert.Str2Uint32(servertype_elem.Text())
	}

	serverid_elem := cfg_root.FindElement("serverid")
	if serverid_elem != nil {
		cfg.ServerId, _ = convert.Str2Uint64(serverid_elem.Text())
	}

	token_elem := cfg_root.FindElement("token")
	if token_elem != nil {
		cfg.Token = token_elem.Text()
	}

	logdir_elem := cfg_root.FindElement("logdir")
	if logdir_elem == nil {
		return nil, errors.New("Log Dir Error")
	}
	cfg.LogDir = logdir_elem.Text()

	loglevel_elem := cfg_root.FindElement("loglevel")
	if loglevel_elem == nil {
		return nil, errors.New("Log Level Error")
	}
	cfg.LogLevel, _ = convert.Str2Int(loglevel_elem.Text())

	sdclient_addr_elem := cfg_root.FindElement("sdclient_addr")
	if sdclient_addr_elem != nil {
		cfg.SDClientAddr = sdclient_addr_elem.Text()
	}

	sdserver_addr_elem := cfg_root.FindElement("sdserver_addr")
	if sdserver_addr_elem != nil {
		cfg.SDServerAddr = sdserver_addr_elem.Text()
	}

	sdclient_url_elem := cfg_root.FindElement("sdclient_url")
	if sdclient_url_elem != nil {
		cfg.SDClientUrl = sdclient_url_elem.Text()
	}

	sdserverurl_elem := cfg_root.FindElement("sdserver_url")
	if sdserverurl_elem != nil {
		cfg.SDServerUrl = sdserverurl_elem.Text()
	}

	return cfg, nil
}

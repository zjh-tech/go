package frame

import (
	"errors"
	"projects/base/convert"
	"projects/base/etree"
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

var GServerCfg *ServerCfg

func init() {
	GServerCfg = &ServerCfg{
		C2SInterListen: "",
		C2SOuterListen: "",
	}
}

func ReadServerCfg(path string) error {
	doc := etree.NewDocument()
	err := doc.ReadFromFile(path)
	if err != nil {
		return err
	}

	//config
	cfg_root := doc.SelectElement("config")
	if cfg_root == nil {
		return errors.New("server_cfg Xml Config Error")
	}

	servicename_elem := cfg_root.FindElement("servicename")
	if servicename_elem != nil {
		GServerCfg.ServiceName = servicename_elem.Text()
	}

	servertype_elem := cfg_root.FindElement("servertype")
	if servertype_elem != nil {
		GServerCfg.ServerType, _ = convert.Str2Uint32(servertype_elem.Text())
	}

	serverid_elem := cfg_root.FindElement("serverid")
	if serverid_elem != nil {
		GServerCfg.ServerId, _ = convert.Str2Uint64(serverid_elem.Text())
	}

	token_elem := cfg_root.FindElement("token")
	if token_elem != nil {
		GServerCfg.Token = token_elem.Text()
	}

	logdir_elem := cfg_root.FindElement("logdir")
	if logdir_elem == nil {
		return errors.New("Log Dir Error")
	}
	GServerCfg.LogDir = logdir_elem.Text()

	loglevel_elem := cfg_root.FindElement("loglevel")
	if loglevel_elem == nil {
		return errors.New("Log Level Error")
	}
	GServerCfg.LogLevel, _ = convert.Str2Int(loglevel_elem.Text())

	sdclient_addr_elem := cfg_root.FindElement("sdclient_addr")
	if sdclient_addr_elem != nil {
		GServerCfg.SDClientAddr = sdclient_addr_elem.Text()
	}

	sdserver_addr_elem := cfg_root.FindElement("sdserver_addr")
	if sdserver_addr_elem != nil {
		GServerCfg.SDServerAddr = sdserver_addr_elem.Text()
	}

	sdclient_url_elem := cfg_root.FindElement("sdclient_url")
	if sdclient_url_elem != nil {
		GServerCfg.SDClientUrl = sdclient_url_elem.Text()
	}

	sdserverurl_elem := cfg_root.FindElement("sdserver_url")
	if sdserverurl_elem != nil {
		GServerCfg.SDServerUrl = sdserverurl_elem.Text()
	}

	return nil
}

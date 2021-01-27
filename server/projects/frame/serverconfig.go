package frame

import (
	"errors"
	"projects/thirds/etree"
	"projects/util"
)

type ServerCfg struct {
	ServiceName string
	serverType  uint32
	ServerId    uint64
	Token       string //ServiceDiscovery的Token
	//日志
	LogDir   string
	LogLevel int

	//Tcp
	SDClientAddr string
	SDServerAddr string

	//Http
	SDClientUrl string
	SDServerUrl string

	//s-s
	C2SInterListen    string
	C2SOuterListen    string
	C2SListenMaxCount int

	//c-s
	C2SHttpsUrl  string
	C2SHttpsCert string
	C2SHttpsKey  string
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
	cfgRoot := doc.SelectElement("config")
	if cfgRoot == nil {
		return errors.New("server_cfg Xml Config Error")
	}

	servicenameElem := cfgRoot.FindElement("servicename")
	if servicenameElem != nil {
		GServerCfg.ServiceName = servicenameElem.Text()
	}

	servertypeElem := cfgRoot.FindElement("servertype")
	if servertypeElem != nil {
		GServerCfg.serverType, _ = util.Str2Uint32(servertypeElem.Text())
	}

	serveridElem := cfgRoot.FindElement("serverid")
	if serveridElem != nil {
		GServerCfg.ServerId, _ = util.Str2Uint64(serveridElem.Text())
	}

	tokenElem := cfgRoot.FindElement("token")
	if tokenElem != nil {
		GServerCfg.Token = tokenElem.Text()
	}

	logdirElem := cfgRoot.FindElement("logdir")
	if logdirElem == nil {
		return errors.New("Log Dir Error")
	}
	GServerCfg.LogDir = logdirElem.Text()

	loglevelElem := cfgRoot.FindElement("loglevel")
	if loglevelElem == nil {
		return errors.New("Log Level Error")
	}
	GServerCfg.LogLevel, _ = util.Str2Int(loglevelElem.Text())

	sdClientAddrElem := cfgRoot.FindElement("sdclient_addr")
	if sdClientAddrElem != nil {
		GServerCfg.SDClientAddr = sdClientAddrElem.Text()
	}

	sdServerAddrElem := cfgRoot.FindElement("sdserver_addr")
	if sdServerAddrElem != nil {
		GServerCfg.SDServerAddr = sdServerAddrElem.Text()
	}

	sdClientUrlElem := cfgRoot.FindElement("sdclient_url")
	if sdClientUrlElem != nil {
		GServerCfg.SDClientUrl = sdClientUrlElem.Text()
	}

	sdserverurlElem := cfgRoot.FindElement("sdserver_url")
	if sdserverurlElem != nil {
		GServerCfg.SDServerUrl = sdserverurlElem.Text()
	}

	return nil
}

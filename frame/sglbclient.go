package frame

import (
	"encoding/json"

	"github.com/zjh-tech/go-frame/engine/enet"
)

type ServiceSpec struct {
	ServiceID   uint64 `json:"service_id"`
	ServiceType uint32 `json:"service_type"`
	Token       string `json:"token"`
	InterAddr   string `json:"inter_addr"`
	OuterAddr   string `json:"out_addr"`
	State       uint32 `json:"state"`
}

type ServiceRegisterReq struct {
	ServiceSpec *ServiceSpec `json:"server_spec"`
}

type ServiceRegisterRes struct {
	Code            int            `json:"code"`
	Message         string         `json:"message"`
	ServiceSpecList []*ServiceSpec `json:"server_spec_list"`
}

type SelectMinServiceReq struct {
	ServiceType uint32 `json:"service_type"`
}

type SelectMinServiceRes struct {
	Code        int          `json:"code"`
	Message     string       `json:"message"`
	ServiceType uint32       `json:"service_type"`
	ServiceSpec *ServiceSpec `json:"server_spec"`
}

const (
	ServiceRegisterOk               int = 0
	ServiceRegisterInputParasFailed int = 1
)

const (
	SelectMinServiceOk        int = 0
	SelectMinINputParasFailed int = 1
	SelectMinServiceFial      int = 2
)

type SGLBClient struct {
	url        string
	httpsFlag  bool
	handlerMap map[string]SGLBClientHandlerFunc
}

func NewSGLBClient() *SGLBClient {
	client := &SGLBClient{
		handlerMap: make(map[string]SGLBClientHandlerFunc),
	}
	return client
}

//-----------------------------------------------------------------------------------------------
func (s *SGLBClient) Init(url string, httpsFlag bool) bool {
	s.url = url
	s.httpsFlag = httpsFlag
	s.RegisterHandler("service_register", OnHandlerServiceRegisterRes)
	return true
}

func (s *SGLBClient) SGLBPost(router string, datas []byte, cbFunc enet.HttpCbFunc, paras interface{}) {
	var fullUrl string
	if s.httpsFlag {
		fullUrl = "https://" + s.url + "/" + router
	} else {
		fullUrl = "http://" + s.url + "/" + router
	}
	content, postErr := enet.Post(fullUrl, datas)
	if postErr != nil {
		ELog.ErrorAf("SGLBPost Error %v", postErr)
		return
	}
	httpEvt := enet.NewHttpEvent(s, router, cbFunc, content, paras)
	enet.GNet.PushSingleHttpEvent(httpEvt)
}

//------------------------------------SLB API-----------------------------------------------------
func (s *SGLBClient) SendServiceRegisterReq(spec *ServiceRegisterReq) {
	datas, marshalErr := json.Marshal(spec)
	if marshalErr != nil {
		ELog.ErrorAf("ServiceRegisterReq json.Marshal Error %v", marshalErr)
		return
	}

	go s.SGLBPost("service_register", datas, nil, nil)
}

func (s *SGLBClient) SendSelectMinServiceReq(serviceType uint32, cbFunc enet.HttpCbFunc, paras interface{}) {
	req := &SelectMinServiceReq{}
	req.ServiceType = serviceType
	datas, marshalErr := json.Marshal(req)
	if marshalErr != nil {
		ELog.ErrorAf("SelectMinServiceReq json.Marshal Error %v", marshalErr)
		return
	}

	go s.SGLBPost("select_min_service", datas, cbFunc, paras)
}

type SGLBClientHandlerFunc func(datas []byte, paras interface{})

func (s *SGLBClient) RegisterHandler(router string, handler SGLBClientHandlerFunc) {
	s.handlerMap[router] = handler
}

func (s *SGLBClient) OnHandler(router string, datas []byte, paras interface{}) {
	if handler, ok := s.handlerMap[router]; ok {
		handler(datas, paras)
	} else {
		ELog.WarnAf("SGLBClient OnHandler Not Find %v Error", router)
	}
}

func OnHandlerServiceRegisterRes(datas []byte, paras interface{}) {
	var res ServiceRegisterRes
	unmarshalErr := json.Unmarshal(datas, &res)
	if unmarshalErr != nil {
		ELog.ErrorAf("ServiceRegisterRes json.Unmarshal Error %v", unmarshalErr)
		return
	}
	ELog.InfoAf("ServiceRegisterRes=%+v", res)
}

var GSGLBClient *SGLBClient

func init() {
	GSGLBClient = NewSGLBClient()
}

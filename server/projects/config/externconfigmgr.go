package config

type ExternConfigMgr struct {
}

func (e *ExternConfigMgr) Load() error {
	return nil
}

var GExternConfigMgr *ExternConfigMgr

func init() {
	GExternConfigMgr = &ExternConfigMgr{}
}

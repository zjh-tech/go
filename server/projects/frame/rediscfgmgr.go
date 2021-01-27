package frame

import (
	"errors"
	"projects/go-engine/eredis"
	"projects/thirds/etree"
	"projects/util"
)

type RedisCfgMgr struct {
	ConnMaxCount   uint64
	RedisConnSpecs []*eredis.RedisConnSpec
}

func NewRedisCfgMgr() *RedisCfgMgr {
	return &RedisCfgMgr{
		RedisConnSpecs: make([]*eredis.RedisConnSpec, 0),
	}
}

func (r *RedisCfgMgr) Load(path string) error {
	doc := etree.NewDocument()
	err := doc.ReadFromFile(path)
	if err != nil {
		return err
	}

	redisElem := doc.SelectElement("redis")
	if redisElem == nil {
		return errors.New("redis_cfg.xml redis Error")
	}

	for _, connectElem := range redisElem.FindElements("connect") {
		connSpec := &eredis.RedisConnSpec{}
		for _, attr := range connectElem.Attr {
			if attr.Key == "Name" {
				connSpec.Name = attr.Value
			}

			if attr.Key == "Host" {
				connSpec.Host = attr.Value
			}

			if attr.Key == "Port" {
				port, _ := util.Str2Int(attr.Value)
				connSpec.Port = port
			}

			if attr.Key == "Password" {
				connSpec.Password = attr.Value
			}
		}
		r.RedisConnSpecs = append(r.RedisConnSpecs, connSpec)
		r.ConnMaxCount++
	}

	//elog.InfoAf("[RedisCfgMgr] ConnMaxCount=%v,RedisConnSpecs=%+v", r.ConnMaxCount, r.RedisConnSpecs)

	return nil
}

var GRedisCfgMgr *RedisCfgMgr

func init() {
	GRedisCfgMgr = NewRedisCfgMgr()
}

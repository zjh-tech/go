package edb

import (
	"errors"

	"github.com/beevik/etree"
	"github.com/zjh-tech/go-frame/base/convert"
)

type DatabaseCfgMgr struct {
	DBTableMaxCount uint64
	DBConnMaxCount  uint64
	DBConnSpecs     []*DBConnSpec
}

func NewDatabaseCfgMgr() *DatabaseCfgMgr {
	return &DatabaseCfgMgr{
		DBTableMaxCount: 0,
		DBConnMaxCount:  0,
		DBConnSpecs:     make([]*DBConnSpec, 0),
	}
}

func (d *DatabaseCfgMgr) Load(path string) error {
	doc := etree.NewDocument()
	err := doc.ReadFromFile(path)
	if err != nil {
		return err
	}

	dbElem := doc.SelectElement("db")
	if dbElem == nil {
		return errors.New("db_cfg.xml db Error")
	}

	for _, dbAttr := range dbElem.Attr {
		if dbAttr.Key == "TableAmount" {
			d.DBTableMaxCount, _ = convert.Str2Uint64(dbAttr.Value)
		}
	}

	if d.DBTableMaxCount == 0 {
		return errors.New("db_cfg.xml TableAmount Error")
	}

	for _, connectElem := range dbElem.FindElements("connect") {
		dbInfo := &DBConnSpec{}
		for _, attr := range connectElem.Attr {
			if attr.Key == "DBHost" {
				dbInfo.Ip = attr.Value
			}

			if attr.Key == "DBPort" {
				port, _ := convert.Str2Uint32(attr.Value)
				dbInfo.Port = uint32(port)
			}

			if attr.Key == "DBUser" {
				dbInfo.User = attr.Value
			}

			if attr.Key == "DBPassword" {
				dbInfo.Password = attr.Value
			}

			if attr.Key == "DBName" {
				dbInfo.Name = attr.Value
			}

			if attr.Key == "Charset" {
				dbInfo.Charset = attr.Value
			}
		}
		d.DBConnSpecs = append(d.DBConnSpecs, dbInfo)
		d.DBConnMaxCount++
	}

	ELog.InfoAf("[DatabaseCfgMgr] DBConnMaxCount=%v DBTableMaxCount=%v,DBConnSpecs=%+v", d.DBConnMaxCount, d.DBTableMaxCount, d.DBConnSpecs)
	return nil
}

var GDatabaseCfgMgr *DatabaseCfgMgr

func init() {
	GDatabaseCfgMgr = NewDatabaseCfgMgr()
}

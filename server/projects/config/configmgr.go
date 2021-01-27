package config
                     
import (
   "projects/go-engine/elog"
   "io/ioutil"
   "os"
   "github.com/golang/protobuf/proto"
)
                     
type ConfigMgr struct {
   DirPath   string
   Tip   Tipcfg
}
                     
func (c *ConfigMgr) LoadCfg(dirPath string, fileName string, message proto.Message) error {
   PbName := dirPath + "/" + fileName + ".pb"
   file, openErr := os.Open(PbName)
   if openErr != nil {
      return openErr
   }
   defer file.Close()
   datas, readErr := ioutil.ReadAll(file)
   if readErr != nil {
      return readErr
   }
   return proto.Unmarshal(datas, message)
}
                     
func (c *ConfigMgr) LoadAllCfg(dirPath string) error {
   c.DirPath   =   dirPath
   if err := c.LoadCfg(dirPath,"tip",&c.Tip); err != nil {
      return err
   }
   elog.InfoA("Tip Config Start")
   for _, info := range c.Tip.Datas {
      elog.InfoAf("tip=%+v \n", info)
   }
   elog.InfoA("Tip Config End")
   return nil
}
                     
func (c *ConfigMgr) GetTipByID(id uint32) *Tip {
   if  c.Tip.Datas != nil {
      if info, ok := c.Tip.Datas[id]; ok {
         return info
      }
   }
   return nil
}

var GConfigMgr *ConfigMgr
                     
func init() {
      GConfigMgr = &ConfigMgr{}
}

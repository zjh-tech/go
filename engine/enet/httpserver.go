package enet

import (
	"fmt"

	"github.com/gin-gonic/gin"
)

type HttpHandlerFunc func(engine *gin.Engine) error

func HttpListen(addr string, cert string, key string, handler HttpHandlerFunc) {
	if addr == "" {
		message := fmt.Sprintf("HttpListen  Addr=%v Empty", addr)
		panic(message)
	}

	go func() {
		//gin.SetMode(gin.ReleaseMode)
		engine := gin.Default()
		handler(engine)

		if len(cert) != 0 && len(key) != 0 {
			ELog.Infof("Https RunTLS %v", addr)
			if err := engine.RunTLS(addr, cert, key); err != nil {
				message := fmt.Sprintf("Https RunTLS Addr=%v Error=%v", addr, err)
				ELog.Infof(message)
				panic(message)
			}
		} else {
			ELog.Infof("Http Run %v", addr)
			if err := engine.Run(addr); err != nil {
				ELog.Infof("Http Run Addr=%v Error=%v", addr, err)
			}
		}
	}()
}

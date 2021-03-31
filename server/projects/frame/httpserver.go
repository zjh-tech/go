package frame

import (
	"net/http"
	"projects/go-engine/elog"
)

func StartHttpsServer(addr string, cert_file string, key_file string, h http.Handler) {
	elog.Infof("Start Https Serve Listen %v", addr)
	if err := http.ListenAndServeTLS(addr, cert_file, key_file, h); err != nil {
		elog.Infof("[HttpsServer] Start Addr=%v Error=%v", addr, err)
		GServer.Quit()
	}
}

func StartHttpServer(addr string, h http.Handler) {
	elog.Infof("Start Http Serve Listen %v", addr)
	if err := http.ListenAndServe(addr, h); err != nil {
		elog.Infof("[HttpServer] Start Addr=%v Error=%v", addr, err)
		GServer.Quit()
	}
}

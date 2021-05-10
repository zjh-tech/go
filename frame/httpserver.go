package frame

import (
	"net/http"
)

func StartHttpsServer(addr string, cert_file string, key_file string, h http.Handler) {
	ELog.Infof("Start Https Serve Listen %v", addr)
	if err := http.ListenAndServeTLS(addr, cert_file, key_file, h); err != nil {
		ELog.Infof("[HttpsServer] Start Addr=%v Error=%v", addr, err)
		GServer.Quit()
	}
}

func StartHttpServer(addr string, h http.Handler) {
	ELog.Infof("Start Http Serve Listen %v", addr)
	if err := http.ListenAndServe(addr, h); err != nil {
		ELog.Infof("[HttpServer] Start Addr=%v Error=%v", addr, err)
		GServer.Quit()
	}
}

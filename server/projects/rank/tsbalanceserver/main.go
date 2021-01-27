package main

func main() {
	var server TsBalanceServer
	if server.Init() {
		server.Run()
	}
	server.UnInit()
}

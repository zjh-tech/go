package main

func main() {
	var server TsRankServer
	if server.Init() {
		server.Run()
	}
	server.UnInit()
}

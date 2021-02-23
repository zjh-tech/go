package main

func main() {
	var server RankServer
	if server.Init() {
		server.Run()
	}
	server.UnInit()
}

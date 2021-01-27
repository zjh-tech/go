package main

func main() {
	var server MatchServer
	if server.Init() {
		server.Run()
	}
	server.UnInit()
}

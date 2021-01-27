package main

func main() {
	var server BattleServer
	if server.Init() {
		server.Run()
	}
	server.UnInit()
}

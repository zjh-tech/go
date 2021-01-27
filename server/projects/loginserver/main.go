package main

func main() {
	var server LoginServer
	if server.Init() {
		server.Run()
	}
	server.UnInit()
}

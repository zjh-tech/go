package main

func main() {
	var server GatewayServer
	if server.Init() {
		server.Run()
	}
	server.UnInit()
}

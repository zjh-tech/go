package main

func main() {
	var server TsGatewayServer
	if server.Init() {
		server.Run()
	}
	server.UnInit()
}

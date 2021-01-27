package main

func main() {
	var server RegistryServer
	if server.Init() {
		server.Run()
	}
	server.UnInit()
}

package main

func main() {
	var server CenterServer
	if server.Init() {
		server.Run()
	}
	server.UnInit()
}

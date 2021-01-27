package main

func main() {
	var server DbServer
	if server.Init() {
		server.Run()
	}
	server.UnInit()
}

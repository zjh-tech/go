package main

func main() {
	var server SGServer
	if server.Init() {
		server.Run()
	}
	server.UnInit()
}

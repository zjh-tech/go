package main

func main() {
	var server HallServer
	if server.Init() {
		server.Run()
	}
	server.UnInit()
}

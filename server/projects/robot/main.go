package main

func main() {
	var robot Robot
	if robot.Init() {
		robot.Run()
	}

	robot.UnInit()
}

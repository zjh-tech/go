package main

func main() {
	var robot TsRobot
	if robot.Init() {
		robot.Run()
	}

	robot.UnInit()
}

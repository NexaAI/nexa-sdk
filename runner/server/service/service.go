package service

func Init() {
	keepAlive.start()
}

func DeInit() {
	keepAlive.stop()
}

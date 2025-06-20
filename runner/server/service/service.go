package service

func Init() {
	keepAlive = NewKeepAlive()
	keepAlive.start()
}

func DeInit() {
	keepAlive.stop()
}

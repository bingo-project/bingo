package bootstrap

func Boot() {
	InitLog()
	InitTimezone()
	InitQueue()
	InitScheduler()
	InitMail()
}

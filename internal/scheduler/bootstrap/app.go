package bootstrap

func Boot() {
	InitLog()
	InitTimezone()
	InitStore()
	InitQueue()
	InitScheduler()
	InitMail()
}

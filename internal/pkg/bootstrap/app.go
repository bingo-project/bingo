package bootstrap

func Boot() {
	InitLog()
	InitTimezone()
	InitSnowflake()
	InitMail()
	InitCache()
	InitAES()
	InitQueue()
}

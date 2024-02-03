package bootstrap

func Boot() {
	InitLog()
	InitJwt()
	InitStore()
	InitCache()
	InitQueue()
	InitTimezone()
	InitAES()
	InitSnowflake()
	InitMail()
}

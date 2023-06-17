package bootstrap

func Boot() {
	InitLog()
	InitJwt()
	InitStore()
	InitCache()
	InitTimezone()
}

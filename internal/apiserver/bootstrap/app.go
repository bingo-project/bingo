package bootstrap

func Boot() {
	InitConfig()
	InitLog()
	InitJwt()
	InitStore()
	InitCache()
}

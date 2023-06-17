package bootstrap

import (
	"github.com/bingo-project/component-base/web/token"

	"bingo/internal/apiserver/config"
)

func InitJwt() {
	// 设置 token 包的签发密钥，用于 token 包 token 的签发和解析
	token.Init(config.Cfg.JWT.SecretKey, config.Cfg.JWT.TTL)
}

package bootstrap

import (
	"github.com/bingo-project/component-base/web/token"

	"github.com/bingo-project/bingo/internal/pkg/facade"
)

func InitJwt() {
	// 设置 token 包的签发密钥，用于 token 包 token 的签发和解析
	token.Init(facade.Config.JWT.SecretKey, facade.Config.JWT.TTL)
}

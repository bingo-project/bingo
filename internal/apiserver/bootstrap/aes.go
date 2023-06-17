package bootstrap

import (
	"github.com/bingo-project/component-base/crypt"

	"bingo/internal/apiserver/facade"
)

func InitAES() {
	facade.AES = crypt.NewAES(facade.Config.Server.Key)
}

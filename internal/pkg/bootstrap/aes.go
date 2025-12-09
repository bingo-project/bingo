package bootstrap

import (
	"github.com/bingo-project/component-base/crypt"

	"github.com/bingo-project/bingo/internal/pkg/facade"
)

func InitAES() {
	facade.AES = crypt.NewAES(facade.Config.App.Key)
}

package middleware

import (
	"os"
	"testing"
	"time"

	"github.com/bingo-project/component-base/cache"
	gocache "github.com/patrickmn/go-cache"
	"github.com/smartystreets/goconvey/convey"

	"github.com/bingo-project/bingo/internal/pkg/config"
	"github.com/bingo-project/bingo/internal/pkg/facade"
	"github.com/bingo-project/bingo/internal/pkg/store"
)

func init() {
	facade.Config.OpenAPI = config.OpenAPI{}
	facade.Cache = cache.NewService(&cache.LocalStore{
		GoCacheClient: gocache.New(time.Minute*5, time.Minute*10),
		KeyPrefix:     "test:cache:",
	})
	store.S = store.NewStore(nil)
}

func TestMain(m *testing.M) {
	// Convey 入口
	convey.SuppressConsoleStatistics()

	result := m.Run()

	// Convey 结果打印
	convey.PrintConsoleStatistics()

	os.Exit(result)
}

package apiserver

import (
	"bingo/internal/apiserver/bootstrap"
)

// run 函数是实际的业务代码入口函数.
func run() error {
	g := bootstrap.InitRouter()

	// 创建并运行 HTTP 服务器
	return startInsecureServer(g)
}

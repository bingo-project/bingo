package admserver

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/bingo-project/component-base/log"
)

// run 函数是实际的业务代码入口函数.
// kill 默认会发送 syscall.SIGTERM 信号
// kill -2 发送 syscall.SIGINT 信号，我们常用的 CTRL + C 就是触发系统 SIGINT 信号
// kill -9 发送 syscall.SIGKILL 信号，但是不能被捕获，所以不需要添加它.
func run() error {
	// 启动 http 服务
	httpServer := NewHttp()
	httpServer.Run()

	// 启动 grpc 服务
	grpcServer := NewGRPC()
	grpcServer.Run()

	// 等待中断信号优雅地关闭服务器（10 秒超时)。
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Infow("Shutting down server ...")

	// 停止服务
	httpServer.Close()
	grpcServer.Close()

	return nil
}

package watcher

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	"bingo/internal/apiserver"
	"bingo/internal/pkg/log"
)

// run 函数是实际的业务代码入口函数.
func run() error {
	// 初始化 store 层
	if err := apiserver.InitStore(); err != nil {
		return err
	}

	cron := newWatchJob().addWatchers()
	cron.Start()

	// 等待中断信号优雅地关闭服务器（10 秒超时)。
	quit := make(chan os.Signal, 1)
	// kill 默认会发送 syscall.SIGTERM 信号
	// kill -2 发送 syscall.SIGINT 信号，我们常用的 CTRL + C 就是触发系统 SIGINT 信号
	// kill -9 发送 syscall.SIGKILL 信号，但是不能被捕获，所以不需要添加它
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM) // 此处不会阻塞
	<-quit                                               // 阻塞在此，当接收到上述两种信号时才会往下执行
	log.Infow("Shutting down server ...")

	ctx := cron.Stop()
	select {
	case <-ctx.Done():
		log.Infow("cron jobs stopped.")
	case <-time.After(3 * time.Minute):
		log.Errorw("context was not done after 3 minutes.")
	}

	return nil
}

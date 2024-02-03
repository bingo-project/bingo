package apiserver

import (
	"github.com/bingo-project/component-base/log"
	"github.com/hibiken/asynq"

	"bingo/internal/apiserver/bootstrap"
	"bingo/internal/apiserver/facade"
	"bingo/internal/apiserver/job"
	"bingo/pkg/queue"
)

// run 函数是实际的业务代码入口函数.
func run() error {
	bootstrap.Boot()

	// 启动 queue worker.
	go runJobs()

	g := bootstrap.InitGin()

	// 创建并运行 HTTP 服务器
	return startInsecureServer(g)
}

func runJobs() {
	mux := asynq.NewServeMux()
	mux.Use(queue.Logging)

	job.AddJobs(mux)

	err := facade.Worker.Run(mux)
	if err != nil {
		log.Fatalw("run worker failed", "err", err)
	}
}

package watcher

import (
	"fmt"

	"github.com/bingo-project/component-base/version/verflag"
	"github.com/spf13/cobra"

	"bingo/internal/apiserver"
	"bingo/internal/apiserver/config"
	"bingo/internal/pkg/log"
)

// NewWatcherCommand creates an App object with default parameters.
func NewWatcherCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "bingo-watcher",
		Short:        "Watcher server",
		Long:         `Watcher is a pluggable watcher service used to do some periodic work like cron job.`,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			verflag.PrintAndExitIfRequested()

			// 初始化日志
			log.Init(config.Cfg.Log)
			defer log.Sync() // Sync 将缓存中的日志刷新到磁盘文件中

			return run()
		},
		// 这里设置命令运行时，不需要指定命令行参数
		Args: func(cmd *cobra.Command, args []string) error {
			for _, arg := range args {
				if len(arg) > 0 {
					return fmt.Errorf("%q does not take any arguments, got %q", cmd.CommandPath(), args)
				}
			}

			return nil
		},
	}

	// 以下设置，使得 InitConfig 函数在每个命令运行时都会被调用以读取配置
	cobra.OnInitialize(apiserver.InitConfig)

	// 在这里您将定义标志和配置设置。

	// Cobra 支持持久性标志(PersistentFlag)，该标志可用于它所分配的命令以及该命令下的每个子命令
	cmd.PersistentFlags().StringVarP(&apiserver.CfgFile, "config", "c", "", "The path to the configuration file. Empty string for no configuration file.")

	// Cobra 也支持本地标志，本地标志只能在其所绑定的命令上使用
	cmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	// 添加 --version 标志
	verflag.AddFlags(cmd.PersistentFlags())

	return cmd
}

// Package apiserver does all the work necessary to create a APIServer.
package apiserver

import (
	"fmt"

	"github.com/bingo-project/component-base/log"
	"github.com/bingo-project/component-base/version/verflag"
	"github.com/spf13/cobra"

	"bingo/internal/apiserver/store"
	"bingo/internal/pkg/bootstrap"
)

// NewAppCommand creates an App object with default parameters.
func NewAppCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "bingo-apiserver",
		Short:        "A good Go practical project",
		Long:         `A good Go practical project, used to create user with basic information.`,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			verflag.PrintAndExitIfRequested()
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
	cobra.OnInitialize(initConfig)

	// 在这里您将定义标志和配置设置。

	// Cobra 支持持久性标志(PersistentFlag)，该标志可用于它所分配的命令以及该命令下的每个子命令
	cmd.PersistentFlags().StringVarP(&bootstrap.CfgFile, "config", "c", "", "The path to the configuration file. Empty string for no configuration file.")

	// Cobra 也支持本地标志，本地标志只能在其所绑定的命令上使用
	cmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	// 添加 --version 标志
	verflag.AddFlags(cmd.PersistentFlags())

	return cmd
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	bootstrap.InitConfig("bingo-apiserver.yaml")
	bootstrap.Boot()
	bootstrap.InitJwt()

	// Init store
	_ = store.NewStore(bootstrap.InitDB())
}

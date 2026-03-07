package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// 版本信息，构建时通过 -ldflags 注入
var (
	Version   = "dev"
	BuildTime = ""
)

var rootCmd = &cobra.Command{
	Use:   "qdev",
	Short: "Q-DEV 项目脚手架工具",
	Long:  `qdev 是一个用于快速创建基于 q-dev 脚手架项目的 CLI 工具。`,
}

func init() {
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "显示版本信息",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("qdev version %s", Version)
		if BuildTime != "" {
			fmt.Printf(" (built at %s)", BuildTime)
		}
		fmt.Println()
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

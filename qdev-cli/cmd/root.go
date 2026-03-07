package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "qdev",
	Short: "Q-DEV 项目脚手架工具",
	Long:  `qdev 是一个用于快速创建基于 q-dev 脚手架项目的 CLI 工具。`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

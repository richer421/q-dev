package cmd

import (
	"fmt"
	"os"

	"q-dev/conf"
	"q-dev/infra/mysql"
	"q-dev/pkg/logger"
	"q-dev/pkg/mcp"

	"github.com/spf13/cobra"
)

var mcpCmd = &cobra.Command{
	Use:   "mcp",
	Short: "Start MCP server for Claude integration",
	Run: func(cmd *cobra.Command, args []string) {
		// Initialize MySQL (required for app layer)
		if err := mysql.Init(conf.C.MySQL); err != nil {
			fmt.Fprintf(os.Stderr, "mysql init failed: %v\n", err)
			os.Exit(1)
		}

		logger.Infof("Starting MCP server...")
		srv := mcp.NewServer()
		if err := srv.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "MCP server error: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(mcpCmd)
}

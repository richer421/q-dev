package cmd

import (
	"fmt"
	"os"

	"q-dev/conf"
	"q-dev/infra/mysql"
	"q-dev/pkg/logger"

	"github.com/spf13/cobra"
)

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Run database migrations",
	Run: func(cmd *cobra.Command, args []string) {
		if err := mysql.Init(conf.C.MySQL); err != nil {
			fmt.Fprintf(os.Stderr, "mysql init failed: %v\n", err)
			os.Exit(1)
		}
		defer mysql.Close()

		logger.Info("Running migration...")
		if err := mysql.Migrate(); err != nil {
			fmt.Fprintf(os.Stderr, "migration failed: %v\n", err)
			os.Exit(1)
		}
		logger.Info("Migration completed!")
	},
}

func init() {
	rootCmd.AddCommand(migrateCmd)
}

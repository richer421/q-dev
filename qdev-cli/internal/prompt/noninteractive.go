package prompt

import (
	"fmt"
	"os"

	"github.com/richer421/qdev-cli/internal/config"
)

// RunNonInteractive runs without interactive prompts (for testing/CI)
func RunNonInteractive(projectName string) *config.Config {
	cfg := &config.Config{}

	cfg.ProjectName = projectName
	cfg.ModuleName = fmt.Sprintf("github.com/%s/%s", os.Getenv("USER"), projectName)
	cfg.Author = os.Getenv("USER")
	cfg.Description = "A project created by qdev"
	cfg.BackendOnly = false

	return cfg
}

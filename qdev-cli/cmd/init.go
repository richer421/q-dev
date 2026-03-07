package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/richer421/qdev-cli/internal/config"
	"github.com/richer421/qdev-cli/internal/git"
	"github.com/richer421/qdev-cli/internal/prompt"
	"github.com/richer421/qdev-cli/internal/template"
	"github.com/spf13/cobra"
)

var (
	initRepoURL  string
	initTag      string
	initGitToken string
	initSSHKey   string
	initForce    bool
	initBackend  bool
	initYes      bool
)

var initCmd = &cobra.Command{
	Use:   "init [project-name]",
	Short: "创建新项目",
	Long:  `从模板仓库创建一个新的 q-dev 项目。`,
	Args:  cobra.MaximumNArgs(1),
	RunE:  runInit,
}

func init() {
	rootCmd.AddCommand(initCmd)

	initCmd.Flags().StringVar(&initRepoURL, "repo", "", "模板仓库地址")
	initCmd.Flags().StringVar(&initTag, "tag", "", "指定 tag 版本")
	initCmd.Flags().StringVar(&initGitToken, "git-token", "", "Git HTTPS 认证 token")
	initCmd.Flags().StringVar(&initSSHKey, "ssh-key", "", "SSH 私钥路径")
	initCmd.Flags().BoolVarP(&initForce, "force", "f", false, "强制覆盖已存在的目录")
	initCmd.Flags().BoolVar(&initBackend, "backend-only", false, "纯后端模式")
	initCmd.Flags().BoolVarP(&initYes, "yes", "y", false, "跳过交互式提示，使用默认值")
}

func runInit(cmd *cobra.Command, args []string) error {
	var projectName string
	if len(args) > 0 {
		projectName = args[0]
	}

	if projectName != "" {
		if err := prompt.ValidateProjectName(projectName); err != nil {
			return err
		}
	}

	var cfg *config.Config
	var err error

	if initYes {
		// 非交互模式
		cfg = prompt.RunNonInteractive(projectName)
	} else {
		// 交互模式
		cfg, err = prompt.Run(projectName)
		if err != nil {
			return err
		}
	}

	if initRepoURL != "" {
		cfg.RepoURL = initRepoURL
	}
	if initTag != "" {
		cfg.Tag = initTag
	}
	if initGitToken != "" {
		cfg.GitToken = initGitToken
	}
	if initSSHKey != "" {
		cfg.SSHKey = initSSHKey
	}
	if initBackend {
		cfg.BackendOnly = true
	}

	targetPath, err := filepath.Abs(cfg.ProjectName)
	if err != nil {
		return err
	}

	if err := prompt.CheckTargetDir(targetPath, initForce); err != nil {
		return err
	}

	fmt.Printf("\n🚀 正在创建项目 %s...\n", cfg.ProjectName)

	fmt.Println("📥 克隆模板仓库...")
	if err := git.Clone(targetPath, git.CloneOptions{
		RepoURL:  cfg.RepoURL,
		Tag:      cfg.Tag,
		GitToken: cfg.GitToken,
		SSHKey:   cfg.SSHKey,
	}); err != nil {
		return fmt.Errorf("克隆失败: %w", err)
	}

	os.RemoveAll(fmt.Sprintf("%s/.git", targetPath))

	if cfg.BackendOnly {
		fmt.Println("🔧 转换为纯后端模式...")
		transform := template.NewTransform(targetPath, true)
		if err := transform.Run(); err != nil {
			return fmt.Errorf("转换失败: %w", err)
		}
	}

	fmt.Println("📝 渲染模板变量...")
	renderer := template.NewRenderer(cfg.ToTemplateData())
	if err := renderer.RenderDir(targetPath); err != nil {
		return fmt.Errorf("渲染失败: %w", err)
	}

	if err := renderer.RenameGoMod(targetPath); err != nil {
		return fmt.Errorf("更新 go.mod 失败: %w", err)
	}

	// 清理不需要的文件
	if err := renderer.CleanUp(targetPath); err != nil {
		return fmt.Errorf("清理文件失败: %w", err)
	}

	fmt.Printf("\n✅ 项目创建成功！\n\n")
	fmt.Printf("📂 项目目录: %s\n", targetPath)
	fmt.Printf("\n接下来:\n")
	fmt.Printf("  cd %s\n", cfg.ProjectName)
	if cfg.BackendOnly {
		fmt.Printf("  make dev\n")
	} else {
		fmt.Printf("  make infra-up\n")
		fmt.Printf("  make dev\n")
	}

	return nil
}

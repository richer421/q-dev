package cmd

import (
	"fmt"
	"os"
	"os/exec"
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

	// 初始化项目
	if err := initializeProject(targetPath, cfg); err != nil {
		return fmt.Errorf("初始化失败: %w", err)
	}

	fmt.Printf("\n✅ 项目创建成功！\n\n")
	fmt.Printf("📂 项目目录: %s\n", targetPath)
	fmt.Printf("\n接下来:\n")
	if !cfg.BackendOnly {
		fmt.Println("  make infra-up   # 启动基础设施 (MySQL/Redis/Kafka)")
	}
	fmt.Println("  make dev       # 启动后端 (http://localhost:8080)")
	if !cfg.BackendOnly {
		fmt.Println("  make fe-dev    # 启动前端 (http://localhost:8000)")
	}
	fmt.Println("")

	return nil
}

// initializeProject 执行项目初始化
func initializeProject(targetPath string, cfg *config.Config) error {
	// 1. 下载后端依赖
	fmt.Println("📦 下载后端依赖...")
	backendDir := filepath.Join(targetPath, "backend")
	if err := runCommand(backendDir, "go", "mod", "download"); err != nil {
		fmt.Printf("⚠️  后端依赖下载失败: %v\n", err)
	}

	// 2. 安装开发工具
	fmt.Println("🔧 安装开发工具 (air, swag, golangci-lint)...")
	if err := runCommand(backendDir, "go", "install", "github.com/air-verse/air@latest"); err != nil {
		fmt.Printf("⚠️  air 安装失败: %v\n", err)
	}
	if err := runCommand(backendDir, "go", "install", "github.com/swaggo/swag/cmd/swag@latest"); err != nil {
		fmt.Printf("⚠️  swag 安装失败: %v\n", err)
	}
	if err := runCommand(backendDir, "go", "install", "github.com/golangci/golangci-lint/cmd/golangci-lint@latest"); err != nil {
		fmt.Printf("⚠️  golangci-lint 安装失败: %v\n", err)
	}

	// 3. 初始化 Git 仓库
	fmt.Println("📋 初始化 Git 仓库...")
	if err := runCommand(targetPath, "git", "init"); err != nil {
		fmt.Printf("⚠️  Git 初始化失败: %v\n", err)
	} else {
		// 添加初始提交
		if err := runCommand(targetPath, "git", "add", "."); err != nil {
			fmt.Printf("⚠️  Git add 失败: %v\n", err)
		}
		if err := runCommand(targetPath, "git", "commit", "-m", "Initial commit"); err != nil {
			fmt.Printf("⚠️  Git commit 失败: %v\n", err)
		}
	}

	// 4. 如果不是纯后端模式，安装前端依赖
	if !cfg.BackendOnly {
		fmt.Println("📦 安装前端依赖...")
		frontendDir := filepath.Join(targetPath, "frontend")
		if _, err := os.Stat(filepath.Join(frontendDir, "package.json")); err == nil {
			if err := runCommand(frontendDir, "pnpm", "install"); err != nil {
				fmt.Printf("⚠️  前端依赖安装失败: %v (请确保已安装 pnpm)\n", err)
			}
		}
	}

	return nil
}

// runCommand 在指定目录执行命令
func runCommand(dir, name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

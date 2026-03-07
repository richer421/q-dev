package prompt

import (
	"fmt"
	"os"

	"github.com/charmbracelet/huh"
	"github.com/richer421/qdev-cli/internal/config"
)

// Run executes the interactive prompt
func Run(projectName string) (*config.Config, error) {
	cfg := &config.Config{}

	// 如果项目名称已通过参数传入，直接使用
	if projectName != "" {
		cfg.ProjectName = projectName
	}

	// 设置默认值
	defaultModule := fmt.Sprintf("github.com/%s/%s", os.Getenv("USER"), cfg.ProjectName)
	if cfg.ProjectName == "" {
		defaultModule = fmt.Sprintf("github.com/%s/my-project", os.Getenv("USER"))
	}
	defaultAuthor := os.Getenv("USER")
	defaultDesc := "A project created by qdev"

	// 预填充默认值，用户直接回车即可
	cfg.ModuleName = defaultModule
	cfg.Author = defaultAuthor
	cfg.Description = defaultDesc

	// 构建表单
	var forms []huh.Field

	// 如果项目名称未传入，添加项目名称输入
	if projectName == "" {
		forms = append(forms, huh.NewInput().
			Title("Project Name").
			Placeholder("my-project").
			Value(&cfg.ProjectName).
			Validate(func(s string) error {
				if s == "" {
					return fmt.Errorf("project name required")
				}
				return nil
			}))
	}

	// 添加其他字段
	forms = append(forms,
		huh.NewInput().
			Title("Go Module").
			Value(&cfg.ModuleName),
		huh.NewInput().
			Title("Author").
			Value(&cfg.Author),
		huh.NewInput().
			Title("Description").
			Value(&cfg.Description),
	)

	// 项目模式选择
	var mode string = "fullstack"
	forms = append(forms,
		huh.NewSelect[string]().
			Title("Project Mode").
			Options(
				huh.NewOption("Full Stack", "fullstack"),
				huh.NewOption("Backend Only", "backend"),
			).
			Value(&mode),
	)

	// 自定义按键绑定，支持 Tab 切换选项
	keyMap := huh.NewDefaultKeyMap()
	keyMap.Select.Next.SetKeys("down", "j", "tab", "right", "l")
	keyMap.Select.Prev.SetKeys("up", "k", "shift+tab", "left", "h")

	// 运行表单
	err := huh.NewForm(
		huh.NewGroup(forms...),
	).WithKeyMap(keyMap).Run()

	if err != nil {
		return nil, err
	}

	cfg.BackendOnly = mode == "backend"

	return cfg, nil
}

// ValidateProjectName validates the project name
func ValidateProjectName(name string) error {
	if name == "" {
		return fmt.Errorf("project name cannot be empty")
	}
	if len(name) > 100 {
		return fmt.Errorf("project name too long")
	}
	for _, r := range name {
		if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_') {
			return fmt.Errorf("project name contains invalid character: %c", r)
		}
	}
	return nil
}

// CheckTargetDir checks if target directory can be used
func CheckTargetDir(path string, force bool) error {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("%s exists but is not a directory", path)
	}
	if !force {
		entries, err := os.ReadDir(path)
		if err != nil {
			return err
		}
		if len(entries) > 0 {
			return fmt.Errorf("directory %s is not empty, use --force to overwrite", path)
		}
	}
	return nil
}

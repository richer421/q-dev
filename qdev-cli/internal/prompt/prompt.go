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

	// 计算默认值
	defaultModule := fmt.Sprintf("github.com/%s/%s", os.Getenv("USER"), cfg.ProjectName)
	if cfg.ProjectName == "" {
		defaultModule = fmt.Sprintf("github.com/%s/my-project", os.Getenv("USER"))
	}
	defaultAuthor := os.Getenv("USER")
	defaultDesc := "A project created by qdev"

	// 构建表单
	var forms []huh.Field

	// 项目名称
	if projectName == "" {
		forms = append(forms, huh.NewInput().
			Title("Project Name").
			Placeholder("my-project").
			Value(&cfg.ProjectName).
			Validate(func(s string) error {
				if s == "" {
					return fmt.Errorf("required")
				}
				return nil
			}))
	}

	// 计算默认数据库名
	defaultDbName := config.GenerateDbName(cfg.ProjectName)

	// Go 模块名（placeholder 显示默认值，Tab 补全）
	forms = append(forms,
		NewAutoFillInput().
			Title("Go Module").
			Placeholder(defaultModule).
			Suggestion(defaultModule).
			Value(&cfg.ModuleName).
			Validate(func(s string) error {
				if s == "" {
					return fmt.Errorf("required")
				}
				return nil
			}),
		NewAutoFillInput().
			Title("Database Name").
			Placeholder(defaultDbName).
			Suggestion(defaultDbName).
			Value(&cfg.DbName),
		NewAutoFillInput().
			Title("Author").
			Placeholder(defaultAuthor).
			Suggestion(defaultAuthor).
			Value(&cfg.Author),
		NewAutoFillInput().
			Title("Description").
			Placeholder(defaultDesc).
			Suggestion(defaultDesc).
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

	// 设置自定义 KeyMap - 把 tab 从 Input.Next 中移除
	keyMap := huh.NewDefaultKeyMap()
	keyMap.Input.Next.SetKeys("enter", "down")  // tab 不��用于 Next
	keyMap.Input.Prev.SetKeys("up", "shift+tab")
	keyMap.Select.Next.SetKeys("down", "j")
	keyMap.Select.Prev.SetKeys("up", "k", "shift+tab")

	// 运行表单
	err := huh.NewForm(
		huh.NewGroup(forms...),
	).WithKeyMap(keyMap).Run()

	if err != nil {
		return nil, err
	}

	cfg.BackendOnly = mode == "backend"

	// 如果用户没有输入，使用默认值
	if cfg.ModuleName == "" {
		cfg.ModuleName = defaultModule
	}
	if cfg.DbName == "" {
		cfg.DbName = defaultDbName
	}
	if cfg.Author == "" {
		cfg.Author = defaultAuthor
	}
	if cfg.Description == "" {
		cfg.Description = defaultDesc
	}

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
	} else {
		// force 模式下删除已存在的目录
		if err := os.RemoveAll(path); err != nil {
			return fmt.Errorf("failed to remove existing directory: %w", err)
		}
	}
	return nil
}

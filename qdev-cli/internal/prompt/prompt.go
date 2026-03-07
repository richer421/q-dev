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

	// 设置默认模块名
	defaultModule := fmt.Sprintf("github.com/%s/%s", os.Getenv("USER"), cfg.ProjectName)
	if cfg.ProjectName == "" {
		defaultModule = fmt.Sprintf("github.com/%s/my-project", os.Getenv("USER"))
	}

	// 构建表单
	var forms []huh.Field

	// 如果项目名称未传入，添加项目名称输入
	if projectName == "" {
		forms = append(forms, huh.NewInput().
			Title("项目名称").
			Placeholder("my-project").
			Value(&cfg.ProjectName).
			Validate(func(s string) error {
				if s == "" {
					return fmt.Errorf("项目名称不能为空")
				}
				return nil
			}))
	}

	// 添加其他字段
	forms = append(forms,
		huh.NewInput().
			Title("Go 模块名").
			Placeholder(defaultModule).
			Value(&cfg.ModuleName).
			Validate(func(s string) error {
				if s == "" {
					return fmt.Errorf("模块名不能为空")
				}
				return nil
			}),
		huh.NewInput().
			Title("作��").
			Placeholder(os.Getenv("USER")).
			Value(&cfg.Author),
		huh.NewInput().
			Title("描述").
			Placeholder("A project created by qdev").
			Value(&cfg.Description),
	)

	// 项目模式选择
	var mode string = "fullstack"
	forms = append(forms,
		huh.NewSelect[string]().
			Title("项目模式").
			Options(
				huh.NewOption("全栈", "fullstack"),
				huh.NewOption("纯后端", "backend"),
			).
			Value(&mode),
	)

	// 自定义按键绑定，支持 Tab 切换选项
	keyMap := huh.NewDefaultKeyMap()
	keyMap.Select.Down.SetKeys("down", "j", "tab")
	keyMap.Select.Up.SetKeys("up", "k", "shift+tab")

	// 运行表单
	err := huh.NewForm(
		huh.NewGroup(forms...),
	).WithKeyMap(keyMap).Run()

	if err != nil {
		return nil, err
	}

	cfg.BackendOnly = mode == "backend"

	// 设置默认值
	if cfg.ModuleName == "" {
		cfg.ModuleName = defaultModule
	}
	if cfg.Author == "" {
		cfg.Author = os.Getenv("USER")
	}
	if cfg.Description == "" {
		cfg.Description = "A project created by qdev"
	}

	return cfg, nil
}

// ValidateProjectName validates the project name
func ValidateProjectName(name string) error {
	if name == "" {
		return fmt.Errorf("项目名称不能为空")
	}
	if len(name) > 100 {
		return fmt.Errorf("项目名称过长")
	}
	for _, r := range name {
		if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_') {
			return fmt.Errorf("项目名称包含非法字符: %c", r)
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
		return fmt.Errorf("%s 已存在且不是目录", path)
	}
	if !force {
		entries, err := os.ReadDir(path)
		if err != nil {
			return err
		}
		if len(entries) > 0 {
			return fmt.Errorf("目录 %s 不为空，使用 --force 强制覆盖", path)
		}
	}
	return nil
}

package prompt

import (
	"fmt"
	"os"

	"github.com/AlecAivazis/survey/v2"
	"github.com/richer421/qdev-cli/internal/config"
)

// Run executes the interactive prompt
func Run(projectName string) (*config.Config, error) {
	cfg := &config.Config{}

	if projectName == "" {
		prompt := &survey.Input{
			Message: "项目名称:",
			Default: "my-project",
		}
		if err := survey.AskOne(prompt, &cfg.ProjectName, survey.WithValidator(survey.Required)); err != nil {
			return nil, err
		}
	} else {
		cfg.ProjectName = projectName
	}

	defaultModule := fmt.Sprintf("github.com/%s/%s", os.Getenv("USER"), cfg.ProjectName)
	modulePrompt := &survey.Input{
		Message: "Go 模块名:",
		Default: defaultModule,
	}
	if err := survey.AskOne(modulePrompt, &cfg.ModuleName, survey.WithValidator(survey.Required)); err != nil {
		return nil, err
	}

	authorPrompt := &survey.Input{
		Message: "作者:",
		Default: os.Getenv("USER"),
	}
	if err := survey.AskOne(authorPrompt, &cfg.Author); err != nil {
		return nil, err
	}

	descPrompt := &survey.Input{
		Message: "描述:",
		Default: "A project created by qdev",
	}
	if err := survey.AskOne(descPrompt, &cfg.Description); err != nil {
		return nil, err
	}

	modePrompt := &survey.Select{
		Message: "项目模式:",
		Options: []string{"全栈", "纯后端"},
		Default: "全栈",
	}
	var mode string
	if err := survey.AskOne(modePrompt, &mode); err != nil {
		return nil, err
	}
	cfg.BackendOnly = mode == "纯后端"

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

package config

import "strings"

// Config holds all configuration for project generation
type Config struct {
	ProjectName string
	ModuleName  string
	DbName      string // 数据库名（从 ProjectName 生成，如 my-project → my_project）
	Author      string
	Description string
	BackendOnly bool

	RepoURL  string
	Tag      string
	GitToken string
	SSHKey   string
	Force    bool
}

// TemplateData holds variables for template rendering
type TemplateData struct {
	ProjectName string
	ModuleName  string
	DbName      string
	Author      string
	Description string
	Year        int
}

// ToTemplateData converts Config to TemplateData
func (c *Config) ToTemplateData() TemplateData {
	return TemplateData{
		ProjectName: c.ProjectName,
		ModuleName:  c.ModuleName,
		DbName:      c.DbName,
		Author:      c.Author,
		Description: c.Description,
		Year:        2026,
	}
}

// GenerateDbName 从项目名生成数据库名（将 - 转换为 _）
func GenerateDbName(projectName string) string {
	return strings.ReplaceAll(projectName, "-", "_")
}

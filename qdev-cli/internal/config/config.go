package config

// Config holds all configuration for project generation
type Config struct {
	ProjectName string
	ModuleName  string
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
	Author      string
	Description string
	Year        int
}

// ToTemplateData converts Config to TemplateData
func (c *Config) ToTemplateData() TemplateData {
	return TemplateData{
		ProjectName: c.ProjectName,
		ModuleName:  c.ModuleName,
		Author:      c.Author,
		Description: c.Description,
		Year:        2026,
	}
}

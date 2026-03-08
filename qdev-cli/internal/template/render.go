package template

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/richer421/qdev-cli/internal/config"
)

type Renderer struct {
	data config.TemplateData
}

func NewRenderer(data config.TemplateData) *Renderer {
	return &Renderer{data: data}
}

func (r *Renderer) RenderDir(root string) error {
	return filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if shouldIgnore(path) {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		if d.IsDir() {
			return nil
		}

		// 只处理文本文件
		if !isTextFile(path) {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("读取文件失败 %s: %w", path, err)
		}

		rendered := r.Render(string(content))

		if string(content) != rendered {
			if err := os.WriteFile(path, []byte(rendered), 0644); err != nil {
				return fmt.Errorf("写入文件失败 %s: %w", path, err)
			}
		}

		return nil
	})
}

// Render 使用正则替换模板变量，避免解析非模板的 {{ 字符
func (r *Renderer) Render(content string) string {
	result := content

	// 替换已知的模板变量
	replacements := map[string]string{
		`{{\.ProjectName}}`:  r.data.ProjectName,
		`{{\.ModuleName}}`:   r.data.ModuleName,
		`{{\.DbName}}`:       r.data.DbName,
		`{{\.Author}}`:       r.data.Author,
		`{{\.Description}}`:  r.data.Description,
		`{{\.Year}}`:         fmt.Sprintf("%d", r.data.Year),
		`{{ .ProjectName }}`: r.data.ProjectName,
		`{{ .ModuleName }}`:  r.data.ModuleName,
		`{{ .DbName }}`:      r.data.DbName,
		`{{ .Author }}`:      r.data.Author,
		`{{ .Description }}`: r.data.Description,
		`{{ .Year }}`:        fmt.Sprintf("%d", r.data.Year),
	}

	for pattern, value := range replacements {
		re := regexp.MustCompile(regexp.QuoteMeta(pattern))
		result = re.ReplaceAllString(result, value)
	}

	// 全局替换：将模板项目的模块名 q-dev 替换为用户的模块名
	// 这会更新所有的 Go import 路径
	if r.data.ModuleName != "q-dev" {
		// 替换 import 语句中的模块路径
		result = regexp.MustCompile(`q-dev(/|")`).ReplaceAllString(result, r.data.ModuleName+"$1")
		// 替换 .gitignore 中的二进制名
		result = regexp.MustCompile(`(?m)^q-dev$`).ReplaceAllString(result, r.data.ProjectName)
	}

	return result
}

// CleanUp 删除不需要的文件��目录
func (r *Renderer) CleanUp(root string) error {
	// 需要删除的文件和目录
	toDelete := []string{
		".github",
		"qdev-cli",
		"README.md",
		"docs",
	}

	for _, name := range toDelete {
		path := filepath.Join(root, name)
		if _, err := os.Stat(path); err == nil {
			if err := os.RemoveAll(path); err != nil {
				return fmt.Errorf("删除 %s 失败: %w", path, err)
			}
		}
	}

	return nil
}

func (r *Renderer) RenameGoMod(root string) error {
	// 查找 go.mod 文件，可能在根目录或 backend 目录
	goModPaths := []string{
		filepath.Join(root, "go.mod"),       // 纯后端模式
		filepath.Join(root, "backend", "go.mod"), // 全栈模式
	}

	for _, goModPath := range goModPaths {
		content, err := os.ReadFile(goModPath)
		if err != nil {
			continue
		}

		lines := strings.Split(string(content), "\n")
		for i, line := range lines {
			if strings.HasPrefix(line, "module ") {
				lines[i] = fmt.Sprintf("module %s", r.data.ModuleName)
				break
			}
		}

		if err := os.WriteFile(goModPath, []byte(strings.Join(lines, "\n")), 0644); err != nil {
			return err
		}
	}

	return nil
}

func shouldIgnore(path string) bool {
	name := filepath.Base(path)
	ignored := []string{".git", ".gitignore", "node_modules", "vendor", ".idea", ".vscode", ".claude", "qdev-cli", ".github", "README.md", "LICENSE"}
	for _, i := range ignored {
		if name == i {
			return true
		}
	}
	return false
}

// isTextFile 判断是否是文本文件
func isTextFile(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	textExts := map[string]bool{
		".go":   true,
		".md":   true,
		".yaml": true,
		".yml":  true,
		".json": true,
		".toml": true,
		".txt":  true,
		".mod":  true,
		".sum":  true,
		".sh":   true,
		".bash": true,
		".zsh":  true,
		".env":  true,
		".ts":   true,
		".tsx":  true,
		".js":   true,
		".jsx":  true,
		".css":  true,
		".scss": true,
		".less": true,
		".html": true,
		".xml":  true,
		".sql":  true,
		".proto": true,
		".dockerfile": true,
		".makefile": true,
		".cfg":  true,
		".conf": true,
		".config": true,
		".example": true,
		".sample": true,
		".template": true,
		".tpl": true,
	}

	// 检查扩展名
	if textExts[ext] {
		return true
	}

	// 检查文件名
	base := strings.ToLower(filepath.Base(path))
	textFiles := map[string]bool{
		"makefile":      true,
		"dockerfile":    true,
		"license":       true,
		"readme":        true,
		"changelog":     true,
		"contributing":  true,
		".gitignore":    true,
		".dockerignore": true,
		".editorconfig": true,
		".eslintrc":     true,
		".prettierrc":   true,
		"air.toml":      true,
	}

	return textFiles[base]
}

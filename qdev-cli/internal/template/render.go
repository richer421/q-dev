package template

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"text/template"

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

		content, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("读取文件失败 %s: %w", path, err)
		}

		rendered, err := r.Render(string(content))
		if err != nil {
			return fmt.Errorf("渲染模板失败 %s: %w", path, err)
		}

		if string(content) != rendered {
			if err := os.WriteFile(path, []byte(rendered), 0644); err != nil {
				return fmt.Errorf("写入文件失败 %s: %w", path, err)
			}
		}

		return nil
	})
}

func (r *Renderer) Render(content string) (string, error) {
	tmpl, err := template.New("template").Parse(content)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, r.data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

func (r *Renderer) RenameGoMod(root string) error {
	goModPath := filepath.Join(root, "go.mod")
	content, err := os.ReadFile(goModPath)
	if err != nil {
		return nil
	}

	lines := strings.Split(string(content), "\n")
	for i, line := range lines {
		if strings.HasPrefix(line, "module ") {
			lines[i] = fmt.Sprintf("module %s", r.data.ModuleName)
			break
		}
	}

	return os.WriteFile(goModPath, []byte(strings.Join(lines, "\n")), 0644)
}

func shouldIgnore(path string) bool {
	name := filepath.Base(path)
	ignored := []string{".git", ".gitignore", "node_modules", "vendor", ".idea", ".vscode", ".claude", "qdev-cli"}
	for _, i := range ignored {
		if name == i {
			return true
		}
	}
	return false
}

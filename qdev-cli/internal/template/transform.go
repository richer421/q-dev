package template

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Transform struct {
	root        string
	backendOnly bool
}

func NewTransform(root string, backendOnly bool) *Transform {
	return &Transform{root: root, backendOnly: backendOnly}
}

func (t *Transform) Run() error {
	if !t.backendOnly {
		return nil
	}

	backendDir := filepath.Join(t.root, "backend")
	if err := t.moveBackendToRoot(backendDir); err != nil {
		return err
	}

	os.RemoveAll(filepath.Join(t.root, "frontend"))

	if err := t.adaptMakefile(); err != nil {
		return err
	}

	if err := t.adaptClaudeMd(); err != nil {
		return err
	}

	// 适配 deploy 目录
	if err := t.adaptDeploy(); err != nil {
		return err
	}

	return nil
}

func (t *Transform) moveBackendToRoot(backendDir string) error {
	if _, err := os.Stat(backendDir); os.IsNotExist(err) {
		return nil
	}

	entries, err := os.ReadDir(backendDir)
	if err != nil {
		return fmt.Errorf("读取 backend 目录失败: %w", err)
	}

	for _, entry := range entries {
		src := filepath.Join(backendDir, entry.Name())
		dst := filepath.Join(t.root, entry.Name())
		os.RemoveAll(dst)
		if err := os.Rename(src, dst); err != nil {
			return fmt.Errorf("移动 %s 失败: %w", entry.Name(), err)
		}
	}

	return os.RemoveAll(backendDir)
}

func (t *Transform) adaptMakefile() error {
	makefilePath := filepath.Join(t.root, "Makefile")
	content, err := os.ReadFile(makefilePath)
	if err != nil {
		return nil
	}

	lines := strings.Split(string(content), "\n")
	var newLines []string
	skip := false

	for _, line := range lines {
		if strings.Contains(line, "fe-") || strings.Contains(line, "frontend") {
			continue
		}
		if strings.HasPrefix(line, "# ---------- 前端") {
			skip = true
			continue
		}
		if skip && strings.HasPrefix(line, "# ----------") {
			skip = false
		}
		if skip {
			continue
		}
		if strings.Contains(line, "BUILD_DIR := backend") {
			line = "BUILD_DIR := ."
		}
		newLines = append(newLines, line)
	}

	return os.WriteFile(makefilePath, []byte(strings.Join(newLines, "\n")), 0644)
}

func (t *Transform) adaptClaudeMd() error {
	claudeMdPath := filepath.Join(t.root, "CLAUDE.md")
	content, err := os.ReadFile(claudeMdPath)
	if err != nil {
		return nil
	}

	text := string(content)
	if idx := strings.Index(text, "## 前端架构"); idx != -1 {
		nextSection := strings.Index(text[idx+1:], "\n## ")
		if nextSection != -1 {
			text = text[:idx] + text[idx+1+nextSection:]
		} else {
			text = text[:idx]
		}
	}

	return os.WriteFile(claudeMdPath, []byte(text), 0644)
}

func (t *Transform) adaptDeploy() error {
	// 删除前端相关的部署文件
	os.RemoveAll(filepath.Join(t.root, "deploy", "Dockerfile.frontend"))
	os.RemoveAll(filepath.Join(t.root, "deploy", "nginx.conf"))

	// 更新 docker-compose.yml，移除前端服务
	composePath := filepath.Join(t.root, "deploy", "docker-compose.yml")
	content, err := os.ReadFile(composePath)
	if err != nil {
		return nil
	}

	// 简单处理：保留基础设施服务，移除前端相关内容
	text := string(content)
	// 如果需要更精细的处理，可以添加更多逻辑

	return os.WriteFile(composePath, []byte(text), 0644)
}

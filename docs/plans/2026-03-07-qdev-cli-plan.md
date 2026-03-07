# qdev-cli Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Build a standalone CLI tool `qdev` that scaffolds new projects from the q-dev template repository.

**Architecture:** CLI uses Cobra for command parsing, go-git for repository cloning, and survey for interactive prompts. Template rendering uses Go's text/template. The tool clones the template repo to a temp directory, transforms it based on mode (full-stack/backend-only), renders variables, and copies to target.

**Tech Stack:** Go, Cobra, go-git, survey, text/template

---

## Task 1: Project Scaffolding

**Files:**
- Create: `qdev-cli/main.go`
- Create: `qdev-cli/go.mod`

**Step 1: Create project directory and initialize Go module**

```bash
mkdir -p qdev-cli/cmd qdev-cli/internal/config qdev-cli/internal/git qdev-cli/internal/prompt qdev-cli/internal/template
cd qdev-cli && go mod init github.com/richer/qdev-cli
```

**Step 2: Install dependencies**

```bash
go get github.com/spf13/cobra@latest
go get github.com/go-git/go-git/v5@latest
go get github.com/AlecAivazis/survey/v2@latest
```

**Step 3: Create main.go**

```go
package main

import "github.com/richer/qdev-cli/cmd"

func main() {
	cmd.Execute()
}
```

**Step 4: Verify build**

Run: `go build -o qdev .`
Expected: No errors

**Step 5: Commit**

```bash
git add qdev-cli/
git commit -m "chore: initialize qdev-cli project structure"
```

---

## Task 2: Root Command

**Files:**
- Create: `qdev-cli/cmd/root.go`

**Step 1: Write root.go**

```go
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "qdev",
	Short: "Q-DEV 项目脚手架工具",
	Long:  `qdev 是一个用于快速创建基于 q-dev 脚手架项目的 CLI 工具。`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
```

**Step 2: Verify**

Run: `go run .`
Expected: Shows help text

**Step 3: Commit**

```bash
git add qdev-cli/cmd/root.go
git commit -m "feat: add root command"
```

---

## Task 3: Config Struct

**Files:**
- Create: `qdev-cli/internal/config/config.go`

**Step 1: Write config.go**

```go
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
```

**Step 2: Commit**

```bash
git add qdev-cli/internal/config/
git commit -m "feat: add config struct"
```

---

## Task 4: Interactive Prompt

**Files:**
- Create: `qdev-cli/internal/prompt/prompt.go`

**Step 1: Write prompt.go**

```go
package prompt

import (
	"fmt"
	"os"

	"github.com/AlecAivazis/survey/v2"
	"github.com/richer/qdev-cli/internal/config"
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
```

**Step 2: Commit**

```bash
git add qdev-cli/internal/prompt/
git commit -m "feat: add interactive prompt"
```

---

## Task 5: Git Clone Module

**Files:**
- Create: `qdev-cli/internal/git/clone.go`

**Step 1: Write clone.go**

```go
package git

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-git/go-git/v5"
	gitconfig "github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"github.com/go-git/go-git/v5/storage/memory"
)

const DefaultRepo = "https://github.com/richer/q-dev"

type CloneOptions struct {
	RepoURL  string
	Tag      string
	GitToken string
	SSHKey   string
}

func Clone(targetDir string, opts CloneOptions) error {
	repoURL := opts.RepoURL
	if repoURL == "" {
		repoURL = DefaultRepo
	}

	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return fmt.Errorf("创建目录失败: %w", err)
	}

	cloneOpts := &git.CloneOptions{
		URL:      repoURL,
		Progress: os.Stdout,
	}

	auth, err := getAuth(repoURL, opts.GitToken, opts.SSHKey)
	if err != nil {
		return fmt.Errorf("认证失败: %w", err)
	}
	if auth != nil {
		cloneOpts.Auth = auth
	}

	if opts.Tag != "" {
		return cloneWithTag(targetDir, repoURL, opts.Tag, auth)
	}

	_, err = git.PlainClone(targetDir, false, cloneOpts)
	if err != nil {
		return fmt.Errorf("克隆仓库失败: %w", err)
	}

	return nil
}

func cloneWithTag(targetDir, repoURL, tag string, auth transport.AuthMethod) error {
	remote := git.NewRemote(memory.NewStorage(), &gitconfig.RemoteConfig{
		Name: "origin",
		URLs: []string{repoURL},
	})

	listOpts := &git.ListOptions{}
	if auth != nil {
		listOpts.Auth = auth
	}

	refs, err := remote.List(listOpts)
	if err != nil {
		return fmt.Errorf("获取远程引用失败: %w", err)
	}

	var tagHash plumbing.Hash
	for _, ref := range refs {
		if ref.Name().Short() == tag {
			tagHash = ref.Hash()
			break
		}
	}

	if tagHash.IsZero() {
		return fmt.Errorf("找不到 tag: %s", tag)
	}

	cloneOpts := &git.CloneOptions{
		URL:           repoURL,
		Progress:      os.Stdout,
		ReferenceName: plumbing.NewTagReferenceName(tag),
		SingleBranch:  true,
		Depth:         1,
	}
	if auth != nil {
		cloneOpts.Auth = auth
	}

	_, err = git.PlainClone(targetDir, false, cloneOpts)
	if err != nil {
		return fmt.Errorf("克隆仓库失败: %w", err)
	}

	return nil
}

func getAuth(repoURL, gitToken, sshKey string) (transport.AuthMethod, error) {
	if sshKey != "" {
		return getSSHAuth(sshKey)
	}

	if gitToken != "" {
		return &http.BasicAuth{
			Username: "git",
			Password: gitToken,
		}, nil
	}

	if len(repoURL) > 0 && repoURL[0] != 'h' {
		defaultKey := filepath.Join(os.Getenv("HOME"), ".ssh", "id_rsa")
		if _, err := os.Stat(defaultKey); err == nil {
			return getSSHAuth(defaultKey)
		}
	}

	return nil, nil
}

func getSSHAuth(keyPath string) (transport.AuthMethod, error) {
	key, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, fmt.Errorf("读取 SSH key 失败: %w", err)
	}

	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		return nil, fmt.Errorf("解析 SSH key 失败: %w", err)
	}

	return &ssh.PublicKeys{
		User:   "git",
		Signer: signer,
	}, nil
}
```

**Step 2: Commit**

```bash
git add qdev-cli/internal/git/
git commit -m "feat: add git clone module with auth support"
```

---

## Task 6: Template Rendering

**Files:**
- Create: `qdev-cli/internal/template/render.go`

**Step 1: Write render.go**

```go
package template

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/richer/qdev-cli/internal/config"
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
	ignored := []string{".git", ".gitignore", "node_modules", "vendor", ".idea", ".vscode"}
	for _, i := range ignored {
		if name == i {
			return true
		}
	}
	return false
}
```

**Step 2: Commit**

```bash
git add qdev-cli/internal/template/
git commit -m "feat: add template rendering module"
```

---

## Task 7: Backend-Only Transform

**Files:**
- Create: `qdev-cli/internal/template/transform.go`

**Step 1: Write transform.go**

```go
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
```

**Step 2: Commit**

```bash
git add qdev-cli/internal/template/transform.go
git commit -m "feat: add backend-only transform"
```

---

## Task 8: Init Command

**Files:**
- Create: `qdev-cli/cmd/init.go`

**Step 1: Write init.go**

```go
package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/richer/qdev-cli/internal/git"
	"github.com/richer/qdev-cli/internal/prompt"
	"github.com/richer/qdev-cli/internal/template"
	"github.com/spf13/cobra"
)

var (
	initRepoURL   string
	initTag       string
	initGitToken  string
	initSSHKey    string
	initForce     bool
	initBackend   bool
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
	initCmd.Flags().BoolVar(&initForce, "force", false, "强制覆盖已存在的目录")
	initCmd.Flags().BoolVar(&initBackend, "backend-only", false, "纯后端模式")
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

	cfg, err := prompt.Run(projectName)
	if err != nil {
		return err
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

	fmt.Printf("\n✅ 项目创建成功！\n\n")
	fmt.Printf("📂 项目目录: %s\n", targetPath)
	fmt.Printf("\n接下来:\n")
	fmt.Printf("  cd %s\n", cfg.ProjectName)
	if cfg.BackendOnly {
		fmt.Printf("  make dev\n")
	} else {
		fmt.Printf("  make infra-up\n")
		fmt.Printf("  make dev\n")
	}

	return nil
}
```

**Step 2: Commit**

```bash
git add qdev-cli/cmd/init.go
git commit -m "feat: add init command"
```

---

## Task 9: Final Build and Test

**Step 1: Build**

```bash
cd qdev-cli
go mod tidy
go build -o qdev .
```

**Step 2: Test help**

Run: `./qdev --help`
Expected: Shows help

Run: `./qdev init --help`
Expected: Shows init flags

**Step 3: Final commit**

```bash
git add qdev-cli/
git commit -m "feat: complete qdev-cli implementation"
```

---

## Summary

| Task | Description |
|------|-------------|
| 1 | Project scaffolding |
| 2 | Root command |
| 3 | Config struct |
| 4 | Interactive prompt |
| 5 | Git clone module |
| 6 | Template rendering |
| 7 | Backend-only transform |
| 8 | Init command |
| 9 | Build and test |

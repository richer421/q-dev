# qdev-cli 设计文档

## 概述

qdev-cli 是一个独立的 CLI 工具，用于快速创建基于 q-dev 脚手架的新项目。类似于 `create-react-app`、`npm create vite` 等前端脚手架工具。

## 目标

- 用户执行 `qdev init my-project` 即可创建一个完整的 q-dev 项目
- 支持全栈模式和纯后端模式
- 支持从 Git 仓库拉取模板，包括私有仓库

## 命令设计

```bash
# 交互式创建
qdev init my-project

# 直接指定模式
qdev init my-project --backend-only           # 纯后端
qdev init my-project --full                   # 全栈（默认）

# 指定模板仓库
qdev init my-project --repo https://github.com/user/q-dev

# 指定版本
qdev init my-project --tag v1.0.0

# 私有仓库认证
qdev init my-project --git-token ghp_xxx      # HTTPS + token
qdev init my-project --ssh-key ~/.ssh/id_rsa  # SSH 认证

# 强制覆盖已存在的目录
qdev init my-project --force
```

## 配置项

交互式问答流程：

1. **项目名称**：`my-project`
2. **模块名**：`github.com/user/my-project`（默认基于项目名推断）
3. **作者**：`Your Name`
4. **描述**：`A brief description`
5. **项目模式**：全栈 / 纯后端

## 项目结构

```
qdev-cli/
├── main.go                 # 入口
├── cmd/
│   ├── root.go             # 根命令
│   └── init.go             # init 子命令
├── internal/
│   ├── git/
│   │   └── clone.go        # Git 克隆逻辑（支持 token/SSH）
│   ├── template/
│   │   ├── render.go       # 模板变量替换
│   │   ├── transform.go    # 模式转换（全栈/纯后端）
│   │   └── ignore.go       # 忽略文件处理
│   └── prompt/
│       └── prompt.go       # 交互式问答
├── go.mod
└── go.sum
```

## 核心流程

```
1. 交互式问答
   - 项目名称、模块名、作者、描述
   - 项目模式（全栈 / 纯后端）

2. Git Clone 到临时目录
   - 使用 --repo / --tag 参数（或默认值）
   - 使用 --git-token / --ssh-key 认证（如需）

3. 模板转换
   - 纯后端模式：去掉 backend/ 外壳，移除 frontend/
   - 适配 Makefile / deploy / CLAUDE.md

4. 模板变量替换
   - 遍历所有文件，替换变量
   - 更新 go.mod 模块名

5. 复制到目标目录，清理临时文件
   - 目标目录已存在则报错（除非 --force）
```

## 模式转换

### 全栈模式

保持原始结构不变：

```
my-project/
├── backend/
├── frontend/
├── Makefile
├── deploy/
└── CLAUDE.md
```

### 纯后端模式

去掉 `backend/` 外壳，内容直接作为根目录：

```
my-project/
├── main.go
├── cmd/
├── conf/
├── http/
├── app/
├── domain/
├── infra/
├── pkg/
├── knowledge/
├── gen/
├── go.mod
├── Makefile           # 移除 fe-* 命令
├── deploy/            # 移除前端相关配置
└── CLAUDE.md          # 移除前端相关文档
```

## 模板变量

| 变量 | 说明 | 示例 |
|------|------|------|
| `{{.ProjectName}}` | 项目名称 | `my-project` |
| `{{.ModuleName}}` | Go 模块名 | `github.com/user/my-project` |
| `{{.Author}}` | 作者 | `Your Name` |
| `{{.Description}}` | 描述 | `A brief description` |
| `{{.Year}}` | 当前年份 | `2026` |

## 技术选型

- **语言**：Go（单二进制分发）
- **CLI 框架**：Cobra
- **Git 操作**：go-git
- **交互式问答**：survey
- **模板渲染**：内置 text/template

## 依赖

```
github.com/spf13/cobra
github.com/go-git/go-git/v5
github.com/AlecAivazis/survey/v2
```

## 发布

- 编译为单二进制文件
- 支持 Linux / macOS / Windows
- 通过 GitHub Releases 分发

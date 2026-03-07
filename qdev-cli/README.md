# qdev-cli

Q-DEV 项目脚手架工具，用于快速创建基于 [q-dev](https://github.com/richer421/q-dev) 脚手架的新项目。

## 安装

```bash
# macOS / Linux (推荐)
curl -fsSL https://github.com/richer421/q-dev/releases/latest/download/install.sh | bash

# 或指定版本
curl -fsSL https://github.com/richer421/q-dev/releases/download/v1.0.0/install.sh | bash
```

<details>
<summary>更多选项</summary>

```bash
# 强制重新安装
curl -fsSL ... | bash -s -- --force

# 安装到指定目录
curl -fsSL ... | bash -s -- --dir ~/bin

# 卸载
curl -fsSL ... | bash -s -- --uninstall

# 查看帮助
curl -fsSL ... | bash -s -- --help
```

</details>

<details>
<summary>手动下载</summary>

从 [Releases](https://github.com/richer421/q-dev/releases) 页面下载对应平台的二进制文件：

| 平台 | 文件 |
|------|------|
| macOS (Intel) | `qdev-darwin-amd64` |
| macOS (Apple Silicon) | `qdev-darwin-arm64` |
| Linux (x64) | `qdev-linux-amd64` |
| Linux (ARM64) | `qdev-linux-arm64` |
| Windows | `qdev-windows-amd64.exe` |

```bash
chmod +x qdev-*
sudo mv qdev-* /usr/local/bin/qdev
```

</details>

## 使用

### 交互��创建

```bash
./qdev init my-project
```

按提示输入：
- 项目名称
- Go 模块名
- 作者
- 描述
- 项目模式（全栈 / 纯后端）

### 命令行参数

```bash
# 纯后端模式
./qdev init my-project --backend-only

# 指定模板仓库
./qdev init my-project --repo https://github.com/user/q-dev

# 指定版本
./qdev init my-project --tag v1.0.0

# 私有��库认证 (HTTPS Token)
./qdev init my-project --git-token ghp_xxx

# 私有仓库认证 (SSH Key)
./qdev init my-project --ssh-key ~/.ssh/id_rsa

# 强制覆盖已存在的目录
./qdev init my-project --force
```

### 完整参数

| 参数 | 说明 |
|------|------|
| `--backend-only` | 纯后端模式，移除 frontend 目录 |
| `--repo` | 模板仓库地址，默认 `https://github.com/richer421/q-dev` |
| `--tag` | 指定 tag 版本 |
| `--git-token` | Git HTTPS 认证 token |
| `--ssh-key` | SSH 私钥路径 |
| `--force` | 强制覆盖已存在的目录 |

## 项目模式

### 全栈模式（默认）

保持原始结构：

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
├── Makefile
└── CLAUDE.md
```

## 模板变量

生成的项目中会替换以下变量：

| 变量 | 说明 |
|------|------|
| `{{.ProjectName}}` | 项目名称 |
| `{{.ModuleName}}` | Go 模块名 |
| `{{.Author}}` | 作者 |
| `{{.Description}}` | 描述 |
| `{{.Year}}` | 当前年份 |

## 开发

```bash
cd qdev-cli
go mod tidy
go build -o qdev .
```

## License

MIT

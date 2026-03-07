# qdev-cli

Q-DEV 项目脚手架工具，用于快速创建基于 [q-dev](https://github.com/richer421/q-dev) 脚手架的新项目。

## 安装

### 从源码构建

```bash
git clone https://github.com/richer421/q-dev.git
cd q-dev/qdev-cli
go build -o qdev .
```

### 直接下载

从 [Releases](https://github.com/richer421/q-dev/releases) 页面下载对应平台的二进制文件。

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

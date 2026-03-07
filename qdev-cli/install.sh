#!/bin/bash

# qdev-cli 安装脚本
# 用法:
#   curl -fsSL https://raw.githubusercontent.com/richer421/q-dev/main/qdev-cli/install.sh | bash
#   curl -fsSL https://raw.githubusercontent.com/richer421/q-dev/main/qdev-cli/install.sh | bash -s -- --help

set -e

REPO="richer421/q-dev"
BINARY="qdev"
INSTALL_DIR="/usr/local/bin"

# 颜色
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

error() {
    echo -e "${RED}[ERROR]${NC} $1"
    exit 1
}

hint() {
    echo -e "${BLUE}[HINT]${NC} $1"
}

# 帮助信息
show_help() {
    echo ""
    echo "  Q-DEV 脚手架工具安装程序"
    echo ""
    echo "用法:"
    echo "  curl -fsSL https://raw.githubusercontent.com/richer421/q-dev/main/qdev-cli/install.sh | bash"
    echo ""
    echo "选项:"
    echo "  -f, --force      强制重新安装（覆盖已有版本）"
    echo "  -v, --version    指定版本号（默认最新版）"
    echo "  -d, --dir        指定安装目录（默认 /usr/local/bin）"
    echo "  -u, --uninstall  卸载"
    echo "  -h, --help       显示帮助"
    echo ""
    echo "示例:"
    echo "  # 安装最新版"
    echo "  curl -fsSL ... | bash"
    echo ""
    echo "  # 强制重新安装"
    echo "  curl -fsSL ... | bash -s -- --force"
    echo ""
    echo "  # 安装指定版本"
    echo "  curl -fsSL ... | bash -s -- --version v1.0.0"
    echo ""
    echo "  # 卸载"
    echo "  curl -fsSL ... | bash -s -- --uninstall"
    echo ""
}

# 解析参数
FORCE=false
UNINSTALL=false
VERSION=""
while [[ $# -gt 0 ]]; do
    case $1 in
        -f|--force)
            FORCE=true
            shift
            ;;
        -v|--version)
            VERSION="$2"
            shift 2
            ;;
        -d|--dir)
            INSTALL_DIR="$2"
            shift 2
            ;;
        -u|--uninstall)
            UNINSTALL=true
            shift
            ;;
        -h|--help)
            show_help
            exit 0
            ;;
        *)
            error "未知参数: $1\n使用 --help 查看帮助"
            ;;
    esac
done

# 检测操作系统
detect_os() {
    case "$(uname -s)" in
        Darwin*)    echo "darwin" ;;
        Linux*)     echo "linux" ;;
        CYGWIN*|MINGW*|MSYS*)    echo "windows" ;;
        *)          error "不支持的操作系统: $(uname -s)" ;;
    esac
}

# 检测架构
detect_arch() {
    case "$(uname -m)" in
        x86_64|amd64)    echo "amd64" ;;
        arm64|aarch64)   echo "arm64" ;;
        *)               error "不支持的架构: $(uname -m)" ;;
    esac
}

# 获取已安装版本
get_installed_version() {
    if command -v ${BINARY} &> /dev/null; then
        # 假设二进制有 --version 输出，如果没有则返回 "unknown"
        local ver=$(${BINARY} --version 2>/dev/null || echo "unknown")
        echo "$ver"
    else
        echo ""
    fi
}

# 获取最新版本
get_latest_version() {
    local version
    version=$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" 2>/dev/null | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
    if [ -z "$version" ]; then
        warn "无法获取最新版本，使用 main 分支"
        echo "main"
    else
        echo "$version"
    fi
}

# 卸载
do_uninstall() {
    local bin_path="${INSTALL_DIR}/${BINARY}"

    if [ ! -f "$bin_path" ]; then
        warn "${BINARY} 未安装在 ${INSTALL_DIR}"
        # 检查是否在其他位置
        if command -v ${BINARY} &> /dev/null; then
            local actual_path=$(command -v ${BINARY})
            info "发现 ${BINARY} 在 ${actual_path}"
            read -p "是否删除? [y/N] " -n 1 -r
            echo
            if [[ $REPLY =~ ^[Yy]$ ]]; then
                rm -f "$actual_path"
                info "已删除"
            fi
        fi
        exit 0
    fi

    info "卸载 ${bin_path}..."
    if [ -w "${INSTALL_DIR}" ]; then
        rm -f "$bin_path"
    else
        sudo rm -f "$bin_path"
    fi
    info "卸载成功"
    exit 0
}

# 检查是否已安装
check_installed() {
    local installed_version=$1
    local target_version=$2

    if [ -n "$installed_version" ] && [ "$FORCE" = false ]; then
        echo ""
        warn "已安装 ${BINARY} (版本: ${installed_version})"
        info "最新版本: ${target_version}"

        if [ "$installed_version" = "$target_version" ]; then
            echo ""
            read -p "已是最新版本，是否重新安装? [y/N] " -n 1 -r
            echo
            if [[ ! $REPLY =~ ^[Yy]$ ]]; then
                info "跳过安装"
                exit 0
            fi
        else
            echo ""
            read -p "是否更新到 ${target_version}? [Y/n] " -n 1 -r
            echo
            if [[ $REPLY =~ ^[Nn]$ ]]; then
                info "跳过更新"
                exit 0
            fi
        fi
        FORCE=true
    fi
}

# 下载二进制
download_binary() {
    local os=$1
    local arch=$2
    local version=$3
    local download_url
    local tmp_dir
    local tmp_file

    tmp_dir=$(mktemp -d)
    tmp_file="${tmp_dir}/${BINARY}"

    if [ "$version" = "main" ]; then
        # 如果没有 release，提示用户从源码构建
        error "没有可用的预编译版本。\n请从源码构建:\n  git clone https://github.com/${REPO}.git\n  cd q-dev/qdev-cli && go build -o qdev ."
    fi

    download_url="https://github.com/${REPO}/releases/download/${version}/${BINARY}-${os}-${arch}"

    info "下载 ${download_url}..."

    if ! curl -fsSL --progress-bar --fail "${download_url}" -o "${tmp_file}"; then
        error "下载失败: ${download_url}\n\n请检查版本是否存在:\n  https://github.com/${REPO}/releases\n\n或手动下载后安装"
    fi

    chmod +x "${tmp_file}"
    echo "${tmp_file}"
}

# 安装
install() {
    local tmp_file=$1
    local bin_path="${INSTALL_DIR}/${BINARY}"

    info "安装到 ${bin_path}..."

    # 确保目录存在
    if [ ! -d "${INSTALL_DIR}" ]; then
        if [ -w "$(dirname ${INSTALL_DIR})" ]; then
            mkdir -p "${INSTALL_DIR}"
        else
            sudo mkdir -p "${INSTALL_DIR}"
        fi
    fi

    if [ -w "${INSTALL_DIR}" ]; then
        mv "${tmp_file}" "${bin_path}"
    else
        sudo mv "${tmp_file}" "${bin_path}"
    fi

    # 清理临时目录
    rm -rf "$(dirname "${tmp_file}")"
}

# 验证安装
verify() {
    if command -v ${BINARY} &> /dev/null; then
        info "安装成功！"
        echo ""
        ${BINARY} --help
        echo ""
        info "使用 '${BINARY} init my-project' 创建新项目"
    else
        warn "${BINARY} 已安装但不在 PATH 中"
        hint "请将 ${INSTALL_DIR} 添加到 PATH，或使用绝对路径:\n  ${INSTALL_DIR}/${BINARY} init my-project"
    fi
}

main() {
    echo ""
    echo "  ╔═══════════════════════════════════════╗"
    echo "  ║     Q-DEV 脚手架工具安装程序          ║"
    echo "  ╚═══════════════════════════════════════╝"
    echo ""

    # 卸载模式
    if [ "$UNINSTALL" = true ]; then
        do_uninstall
    fi

    local os arch version installed_version

    os=$(detect_os)
    arch=$(detect_arch)
    installed_version=$(get_installed_version)

    if [ -n "$VERSION" ]; then
        version="$VERSION"
    else
        version=$(get_latest_version)
    fi

    info "操作系统: ${os}"
    info "架构: ${arch}"
    info "目标版本: ${version}"

    if [ -n "$installed_version" ]; then
        info "当前版本: ${installed_version}"
    fi
    echo ""

    # 检查是否已安装
    check_installed "$installed_version" "$version"

    local tmp_file
    tmp_file=$(download_binary "${os}" "${arch}" "${version}")
    install "${tmp_file}"
    verify
}

main "$@"

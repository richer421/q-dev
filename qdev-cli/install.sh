#!/bin/bash

# qdev-cli 安装脚本
# 用法: curl -fsSL https://raw.githubusercontent.com/richer421/q-dev/main/qdev-cli/install.sh | bash

set -e

REPO="richer421/q-dev"
BINARY="qdev"
INSTALL_DIR="/usr/local/bin"

# 颜色
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
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

# 获取最新版本
get_latest_version() {
    local version
    version=$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
    if [ -z "$version" ]; then
        warn "无法获取最新版本，使用 main 分支"
        echo "main"
    else
        echo "$version"
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
        # 从 main 分支下载预编译二进制（如果有）
        download_url="https://github.com/${REPO}/releases/download/latest/${BINARY}-${os}-${arch}"
        info "从 main 分支下载..."
    else
        download_url="https://github.com/${REPO}/releases/download/${version}/${BINARY}-${os}-${arch}"
        info "下载版本 ${version}..."
    fi

    if ! curl -fsSL --fail "${download_url}" -o "${tmp_file}"; then
        error "下载失败: ${download_url}\n请检查版本是否存在，或手动从 GitHub Releases 下载"
    fi

    chmod +x "${tmp_file}"
    echo "${tmp_file}"
}

# 安装
install() {
    local tmp_file=$1

    info "安装到 ${INSTALL_DIR}/${BINARY}..."

    if [ -w "${INSTALL_DIR}" ]; then
        mv "${tmp_file}" "${INSTALL_DIR}/${BINARY}"
    else
        info "需要 sudo 权限安装到 ${INSTALL_DIR}"
        sudo mv "${tmp_file}" "${INSTALL_DIR}/${BINARY}"
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
        error "安装验证失败"
    fi
}

main() {
    echo ""
    echo "  Q-DEV 脚手架工具安装程序"
    echo ""

    local os arch version tmp_file

    os=$(detect_os)
    arch=$(detect_arch)
    version=$(get_latest_version)

    info "操作系统: ${os}"
    info "架构: ${arch}"
    info "版本: ${version}"
    echo ""

    tmp_file=$(download_binary "${os}" "${arch}" "${version}")
    install "${tmp_file}"
    verify
}

main "$@"

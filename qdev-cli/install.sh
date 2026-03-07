#!/bin/bash

# qdev-cli 安装脚本
# 用法:
#   curl -fsSL https://github.com/richer421/q-dev/releases/latest/download/install.sh | bash

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
    echo -e "${GREEN}[INFO]${NC} $1" >&2
}

warn() {
    echo -e "${YELLOW}[WARN]${NC} $1" >&2
}

error() {
    echo -e "${RED}[ERROR]${NC} $1" >&2
    exit 1
}

hint() {
    echo -e "${BLUE}[HINT]${NC} $1" >&2
}

# 帮助信息
show_help() {
    cat << 'EOF'

  Q-DEV 脚手架工具安装程序

用法:
  curl -fsSL https://github.com/richer421/q-dev/releases/latest/download/install.sh | bash

选项:
  -f, --force      强制重新安装
  -v, --version    指定版本号（默认最新版）
  -d, --dir        指定安装目录（默认 /usr/local/bin）
  -u, --uninstall  卸载
  -h, --help       显示帮助

示例:
  curl -fsSL ... | bash                           # 安装最新版
  curl -fsSL ... | bash -s -- --force             # 强制重新安装
  curl -fsSL ... | bash -s -- --version v1.0.0    # 安装指定版本
  curl -fsSL ... | bash -s -- --uninstall         # 卸载

EOF
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
        ${BINARY} version 2>/dev/null | head -1 || echo "unknown"
    fi
}

# 获取最新版本
get_latest_version() {
    local version
    version=$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" 2>/dev/null | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
    if [ -z "$version" ]; then
        error "无法获取最新版本，请检查网络连接"
    fi
    echo "$version"
}

# 卸载
do_uninstall() {
    local bin_path="${INSTALL_DIR}/${BINARY}"

    if [ ! -f "$bin_path" ]; then
        if command -v ${BINARY} &> /dev/null; then
            local actual_path=$(command -v ${BINARY})
            info "发现 ${BINARY} 在 ${actual_path}"
            read -p "是否删除? [y/N] " -n 1 -r
            echo
            if [[ $REPLY =~ ^[Yy]$ ]]; then
                rm -f "$actual_path"
                info "已删除"
            fi
        else
            warn "${BINARY} 未安装"
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

    TMP_DIR=$(mktemp -d)
    TMP_FILE="${TMP_DIR}/${BINARY}"

    local download_url="https://github.com/${REPO}/releases/download/${version}/${BINARY}-${os}-${arch}"

    info "下载 ${download_url}..."

    if ! curl -fsSL --progress-bar --fail "${download_url}" -o "${TMP_FILE}"; then
        rm -rf "${TMP_DIR}"
        error "下载失败: ${download_url}\n\n请检查版本是否存在:\n  https://github.com/${REPO}/releases"
    fi

    chmod +x "${TMP_FILE}"
}

# 安装
install() {
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
        mv "${TMP_FILE}" "${bin_path}"
    else
        sudo mv "${TMP_FILE}" "${bin_path}"
    fi

    # 清理临时目录
    rm -rf "${TMP_DIR}"
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
        hint "请将 ${INSTALL_DIR} 添加到 PATH"
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
    [ -n "$installed_version" ] && info "当前版本: ${installed_version}"
    echo ""

    # 检查是否已安装
    check_installed "$installed_version" "$version"

    download_binary "${os}" "${arch}" "${version}"
    install
    verify
}

main "$@"

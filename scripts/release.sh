#!/bin/bash

set -e

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 显示帮助信息
show_help() {
    cat << EOF
${GREEN}=== Tracely 发版本脚本 ===${NC}

用法：$0 [选项]

选项:
  -h, --help              显示帮助信息
  -p, --patch             升级修订版本号 (默认，如 1.0.0 -> 1.0.1)
  -m, --minor             升级次版本号 (如 1.0.0 -> 1.1.0)
  -M, --major             升级主版本号 (如 1.0.0 -> 2.0.0)
  -v, --version <version> 指定具体版本号 (如 1.2.3)
  -f, --force             跳过确认提示
  -n, --dry-run           仅显示将要执行的操作，不实际创建 tag

示例:
  $0                      # 自动升级修订版本号 (patch)
  $0 --minor              # 升级次版本号
  $0 --major              # 升级主版本号
  $0 --version 2.0.0      # 指定版本号为 2.0.0
  $0 -p -f                # 升级修订版本号并跳过确认
  $0 --dry-run            # 预览操作

EOF
}

# 默认参数
UPGRADE_TYPE="patch"
SPECIFIC_VERSION=""
SKIP_CONFIRM=false
DRY_RUN=false

# 解析命令行参数
while [[ $# -gt 0 ]]; do
    case $1 in
        -h|--help)
            show_help
            exit 0
            ;;
        -p|--patch)
            UPGRADE_TYPE="patch"
            shift
            ;;
        -m|--minor)
            UPGRADE_TYPE="minor"
            shift
            ;;
        -M|--major)
            UPGRADE_TYPE="major"
            shift
            ;;
        -v|--version)
            SPECIFIC_VERSION="$2"
            shift 2
            ;;
        -f|--force)
            SKIP_CONFIRM=true
            shift
            ;;
        -n|--dry-run)
            DRY_RUN=true
            shift
            ;;
        *)
            echo -e "${RED}错误：未知选项 $1${NC}"
            echo -e "使用 -h 或 --help 查看帮助信息"
            exit 1
            ;;
    esac
done

echo -e "${GREEN}=== Tracely 发版本脚本 ===${NC}"

# 检查是否在 git 仓库中
if ! git rev-parse --git-dir > /dev/null 2>&1; then
    echo -e "${RED}错误：当前目录不是 git 仓库${NC}"
    exit 1
fi

# 检查是否配置了 git remote
if ! git remote | grep -q "origin"; then
    echo -e "${RED}错误：未配置 git remote origin${NC}"
    exit 1
fi

# 获取当前最新的 tag 版本
get_latest_tag() {
    git tag --sort=-v:refname | head -n 1
}

# 解析版本号并升级
upgrade_version() {
    local current_version=$1
    local upgrade_type=$2
    
    if [ -z "$current_version" ]; then
        echo "1.0.0"
        return
    fi
    
    # 移除 'v' 前缀
    local version=${current_version#v}
    
    # 解析主版本、次版本、修订版本
    IFS='.' read -r major minor patch <<< "$version"
    
    # 验证版本号格式
    if ! [[ "$major" =~ ^[0-9]+$ ]] || ! [[ "$minor" =~ ^[0-9]+$ ]] || ! [[ "$patch" =~ ^[0-9]+$ ]]; then
        echo -e "${RED}错误：无效的版本号格式：$current_version${NC}"
        exit 1
    fi
    
    # 根据类型升级版本号
    case $upgrade_type in
        major)
            major=$((major + 1))
            minor=0
            patch=0
            ;;
        minor)
            minor=$((minor + 1))
            patch=0
            ;;
        patch)
            patch=$((patch + 1))
            ;;
    esac
    
    echo "${major}.${minor}.${patch}"
}

# 验证指定的版本号格式
validate_version() {
    local version=$1
    if ! [[ "$version" =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
        echo -e "${RED}错误：无效的版本号格式：$version (应为 X.Y.Z 格式)${NC}"
        exit 1
    fi
}

# 获取当前分支
current_branch=$(git rev-parse --abbrev-ref HEAD)
echo -e "${YELLOW}当前分支：${current_branch}${NC}"

# 检查是否有未提交的更改
if [ -n "$(git status --porcelain)" ]; then
    echo -e "${RED}错误：存在未提交的更改，请先提交或暂存${NC}"
    git status --short
    exit 1
fi

# 获取当前最新的 tag
latest_tag=$(get_latest_tag)
echo -e "${YELLOW}当前最新 tag: ${latest_tag:-无}${NC}"

# 计算新版本号
if [ -n "$SPECIFIC_VERSION" ]; then
    # 使用指定的版本号
    validate_version "$SPECIFIC_VERSION"
    new_version="$SPECIFIC_VERSION"
    echo -e "${BLUE}使用指定版本号：${new_version}${NC}"
else
    # 自动升级版本号
    new_version=$(upgrade_version "$latest_tag" "$UPGRADE_TYPE")
    echo -e "${BLUE}升级类型：${UPGRADE_TYPE}${NC}"
fi

new_tag="v${new_version}"

echo -e "${GREEN}即将创建新版本：${new_tag}${NC}"

# Dry run 模式
if [ "$DRY_RUN" = true ]; then
    echo -e "${YELLOW}[Dry Run] 以下操作将被执行:${NC}"
    echo -e "  1. git pull origin ${current_branch} --tags"
    echo -e "  2. git tag -a ${new_tag} -m \"Release ${new_tag}\""
    echo -e "  3. git push origin ${new_tag}"
    exit 0
fi

# 确认是否继续
if [ "$SKIP_CONFIRM" = false ]; then
    read -p "确认创建 tag ${new_tag} 并推送到 GitHub? (y/n): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        echo -e "${YELLOW}操作已取消${NC}"
        exit 0
    fi
fi

# 拉取最新代码
echo -e "${YELLOW}正在拉取最新代码...${NC}"
git pull origin "${current_branch}" --tags

# 创建 tag
echo -e "${GREEN}创建 tag: ${new_tag}${NC}"
git tag -a "${new_tag}" -m "Release ${new_tag}"

# 推送 tag
echo -e "${GREEN}推送 tag 到 GitHub...${NC}"
git push origin "${new_tag}"

echo -e "${GREEN}✓ 成功创建并推送 tag: ${new_tag}${NC}"
echo -e "${YELLOW}GitHub Actions 将自动构建并发布${NC}"
echo -e "${YELLOW}查看进度：https://github.com/hanxi/tracely/actions${NC}"
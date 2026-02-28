#!/bin/bash

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

echo_info() { echo -e "${GREEN}[INFO]${NC} $1"; }
echo_warn() { echo -e "${YELLOW}[WARN]${NC} $1"; }
echo_error() { echo -e "${RED}[ERROR]${NC} $1"; }
echo_blue() { echo -e "${BLUE}[STEP]${NC} $1"; }

show_help() {
    cat << EOF
Usage: $0 [OPTIONS] [VERSION]

自动打 tag 并推送，触发 GitHub Actions 发布流程

OPTIONS:
    -h, --help          显示帮助信息
    -v, --version       指定版本号（如：1.2.0）
    --major             增加主版本号
    --minor             增加次版本号
    --patch             增加修订版本号
    --dry-run           模拟运行
    --wait              等待 GitHub Actions 完成

EXAMPLES:
    $0                  # 自动检测下一个版本号
    $0 1.2.0            # 指定版本号
    $0 --minor          # 增加 minor 版本
    $0 --patch --wait   # 增加 patch 版本并等待发布完成
EOF
}

check_dependencies() {
    echo_blue "检查依赖..."
    if ! command -v git &> /dev/null; then
        echo_error "缺少依赖：git"
        exit 1
    fi
    if ! command -v gh &> /dev/null; then
        echo_warn "GitHub CLI (gh) 未安装，部分功能将不可用"
    fi
    echo_info "依赖检查通过"
}

check_git_status() {
    echo_blue "检查 Git 状态..."
    local current_branch=$(git rev-parse --abbrev-ref HEAD)
    if [ "$current_branch" != "main" ]; then
        echo_warn "当前不在 main 分支：$current_branch"
        read -p "是否继续？(y/N): " confirm
        [[ "$confirm" != "y" && "$confirm" != "Y" ]] && exit 0
    fi
    if ! git diff-index --quiet HEAD --; then
        echo_warn "存在未提交的更改"
        git status --short
        read -p "是否继续？(y/N): " confirm
        [[ "$confirm" != "y" && "$confirm" != "Y" ]] && exit 0
    fi
    echo_info "Git 状态检查通过"
}

get_latest_version() {
    local latest_tag=$(git tag -l 'v*' --sort=-v:refname | head -n 1)
    [ -n "$latest_tag" ] && echo "${latest_tag#v}" || echo "0.0.0"
}

calculate_next_version() {
    local current_version=$1 bump_type=$2
    IFS='.' read -r major minor patch <<< "$current_version"
    case $bump_type in
        major) echo "$((major + 1)).0.0" ;;
        minor) echo "${major}.$((minor + 1)).0" ;;
        *) echo "${major}.${minor}.$((patch + 1))" ;;
    esac
}

generate_changelog() {
    local version=$1 prev_version=$2
    echo_blue "生成 CHANGELOG..."
    local changelog_file="$PROJECT_ROOT/CHANGELOG.md"
    local date=$(date +%Y-%m-%d)
    local prev_tag="v${prev_version}"
    local changelog_entry="## v${version} (${date})\n\n"
    
    if git rev-parse "$prev_tag" &> /dev/null; then
        changelog_entry+=$(git log --pretty=format:"- %s (%h)" "${prev_tag}..HEAD" 2>/dev/null || echo "- Initial release")
    else
        changelog_entry+=$(git log --pretty=format:"- %s (%h)" --max-count=20 2>/dev/null || echo "- Initial release")
    fi
    changelog_entry+="\n\n"
    
    if [ -f "$changelog_file" ]; then
        local temp_file=$(mktemp)
        echo -e "$changelog_entry" > "$temp_file"
        cat "$changelog_file" >> "$temp_file"
        mv "$temp_file" "$changelog_file"
    else
        echo -e "# Changelog\n\n${changelog_entry}" > "$changelog_file"
    fi
    echo_info "CHANGELOG 已更新：$changelog_file"
}

create_and_push_tag() {
    local version=$1 dry_run=$2
    local tag_name="v${version}"
    echo_blue "创建 tag: ${tag_name}..."
    
    if [ "$dry_run" = "true" ]; then
        echo_warn "[模拟运行] git tag -a ${tag_name} -m \"Release ${tag_name}\""
        echo_warn "[模拟运行] git push origin ${tag_name}"
        return 0
    fi
    
    git tag -a "${tag_name}" -m "Release ${tag_name}"
    echo_blue "推送 tag 到远程仓库..."
    git push origin "${tag_name}"
    echo_info "Tag ${tag_name} 已创建并推送"
}

wait_for_release() {
    local version=$1 tag_name="v${version}"
    command -v gh &> /dev/null || { echo_warn "gh CLI 未安装"; return; }
    
    echo_blue "等待 GitHub Actions 完成..."
    echo_info "Release 页面：https://github.com/hanxi/tracely/releases/tag/${tag_name}"
    
    local max_attempts=60 attempt=0 sleep_time=30
    while [ $attempt -lt $max_attempts ]; do
        attempt=$((attempt + 1))
        echo_info "检查发布状态... (尝试 $attempt/$max_attempts, 等待 ${sleep_time}s)"
        sleep $sleep_time
        if gh release view "${tag_name}" --repo hanxi/tracely &> /dev/null; then
            echo_info "✅ Release 已发布！"
            echo_info "URL: https://github.com/hanxi/tracely/releases/tag/${tag_name}"
            return 0
        fi
    done
    echo_warn "等待超时，请手动检查：https://github.com/hanxi/tracely/actions"
}

main() {
    echo_info "========================================="
    echo_info "  Tracely Auto Release"
    echo_info "========================================="
    echo
    
    local specified_version="" bump_type="" dry_run="false" wait_for_completion="false"
    
    while [[ $# -gt 0 ]]; do
        case $1 in
            -h|--help) show_help; exit 0 ;;
            -v|--version) specified_version="$2"; shift 2 ;;
            --major) bump_type="major"; shift ;;
            --minor) bump_type="minor"; shift ;;
            --patch) bump_type="patch"; shift ;;
            --dry-run) dry_run="true"; shift ;;
            --wait) wait_for_completion="true"; shift ;;
            *) [[ $1 =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]] && specified_version="$1"; shift ;;
        esac
    done
    
    check_dependencies
    check_git_status
    
    local current_version=$(get_latest_version)
    echo_info "当前最新版本：$current_version"
    
    local new_version=""
    if [ -n "$specified_version" ]; then
        new_version="$specified_version"
    elif [ -n "$bump_type" ]; then
        new_version=$(calculate_next_version "$current_version" "$bump_type")
    else
        new_version=$(calculate_next_version "$current_version" "patch")
    fi
    
    echo_info "新版本号：$new_version"
    echo
    
    if [ "$dry_run" != "true" ]; then
        read -p "确认创建并推送 tag v${new_version}？(y/N): " confirm
        [[ "$confirm" != "y" && "$confirm" != "Y" ]] && exit 0
    fi
    
    if [ "$dry_run" != "true" ]; then
        generate_changelog "$new_version" "$current_version"
        if [ -f "$PROJECT_ROOT/CHANGELOG.md" ]; then
            echo_blue "提交 CHANGELOG..."
            git add CHANGELOG.md
            git commit -m "docs: update CHANGELOG for v${new_version}" || true
        fi
    fi
    
    create_and_push_tag "$new_version" "$dry_run"
    
    echo
    echo_info "========================================="
    [ "$dry_run" = "true" ] && echo_info "  模拟运行完成" || echo_info "  Tag 已创建并推送"
    echo_info "========================================="
    echo
    
    if [ "$wait_for_completion" = "true" ] && [ "$dry_run" != "true" ]; then
        wait_for_release "$new_version"
    else
        echo_info "GitHub Actions 将自动触发"
        echo_info "查看进度：https://github.com/hanxi/tracely/actions"
    fi
}

main "$@"

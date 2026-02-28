#!/bin/bash

set -e

# 项目名称
PROJECT_NAME="tracely"
VERSION=${VERSION:-"1.0.0"}
GITHUB_TOKEN=${GITHUB_TOKEN:-""}
REPO=${REPO:-"hanxi/tracely"}

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

echo_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

echo_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 检查依赖
check_dependencies() {
    echo_info "Checking dependencies..."
    
    if ! command -v gh &> /dev/null; then
        echo_warn "GitHub CLI (gh) not found. Installing..."
        if [[ "$OSTYPE" == "darwin"* ]]; then
            brew install gh
        elif [[ "$OSTYPE" == "linux-gnu"* ]]; then
            sudo apt-get update && sudo apt-get install -y gh
        else
            echo_error "Please install GitHub CLI manually: https://github.com/cli/cli#installation"
            exit 1
        fi
    fi
    
    if ! command -v zip &> /dev/null; then
        echo_error "zip not found. Please install zip."
        exit 1
    fi
    
    if ! command -v tar &> /dev/null; then
        echo_error "tar not found. Please install tar."
        exit 1
    fi
    
    echo_info "All dependencies are installed."
}

# 检查 GitHub 登录状态
check_github_login() {
    echo_info "Checking GitHub login status..."
    
    if ! gh auth status &> /dev/null; then
        echo_warn "Not logged in to GitHub. Please login..."
        gh auth login
    fi
    
    echo_info "GitHub login verified."
}

# 创建 GitHub Release
create_release() {
    local TAG_NAME="v${VERSION}"
    local RELEASE_NOTES="./releases/RELEASE_NOTES.md"
    
    echo_info "Creating GitHub release: ${TAG_NAME}..."
    
    # 生成发布说明
    cat > "$RELEASE_NOTES" << EOF
# ${PROJECT_NAME} v${VERSION}

## 更新内容

- 初始版本发布
- 支持错误收集和用户活跃统计
- 提供 Go SDK 和 TypeScript SDK
- 内嵌 Dashboard，单二进制部署

## 下载

选择适合你平台的压缩包：

- **Linux**: \`tracely-${VERSION}-linux-amd64.tar.gz\` 或 \`tracely-${VERSION}-linux-arm64.tar.gz\`
- **macOS**: \`tracely-${VERSION}-darwin-amd64.tar.gz\` 或 \`tracely-${VERSION}-darwin-arm64.tar.gz\`
- **Windows**: \`tracely-${VERSION}-windows-amd64.zip\` 或 \`tracely-${VERSION}-windows-arm64.zip\`

## 快速开始

### 1. 解压

```bash
# Linux/macOS
tar -xzf tracely-${VERSION}-linux-amd64.tar.gz

# Windows
unzip tracely-${VERSION}-windows-amd64.zip
```

### 2. 配置

生成密码哈希：
\`\`\`bash
./hashpwd yourpassword
\`\`\`

编辑 \`config.yaml\`，将生成的密码哈希填入 \`users[].passwordHash\` 字段。

### 3. 运行

```bash
./tracely
```

### 4. 访问 Dashboard

打开浏览器访问：http://localhost:3001

用户名：\`admin\`，密码：你设置的密码。

## 文档

详细文档请查看：https://github.com/${REPO}/blob/main/README.md
EOF
    
    # 创建 Release
    gh release create "$TAG_NAME" \
        --repo "$REPO" \
        --title "${PROJECT_NAME} v${VERSION}" \
        --notes-file "$RELEASE_NOTES" \
        --draft \
        ./releases/${PROJECT_NAME}-*.tar.gz \
        ./releases/${PROJECT_NAME}-*.zip \
        ./releases/SHA256SUMS.txt
    
    echo_info "========================================="
    echo_info "  Release created successfully!"
    echo_info "========================================="
    echo_info ""
    echo_info "Release URL: https://github.com/${REPO}/releases/tag/${TAG_NAME}"
    echo_info ""
    echo_warn "Note: Release is created as DRAFT. Please review and publish it manually."
}

# 主函数
main() {
    echo_info "========================================="
    echo_info "  ${PROJECT_NAME} Release Publisher"
    echo_info "  Version: ${VERSION}"
    echo_info "  Repo: ${REPO}"
    echo_info "========================================="
    
    # 检查依赖
    check_dependencies
    
    # 检查 GitHub 登录
    check_github_login
    
    # 检查构建产物是否存在
    if [ ! -d "./releases" ]; then
        echo_error "Release files not found. Please run ./scripts/release.sh first."
        exit 1
    fi
    
    # 创建 Release
    create_release
}

# 执行
main "$@"

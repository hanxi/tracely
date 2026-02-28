#!/bin/bash

set -e

# 项目名称
PROJECT_NAME="tracely"
VERSION=${VERSION:-"1.0.0"}

# 获取构建时间和 Git 提交哈希
BUILD_TIME=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
GIT_COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

echo_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

echo_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 清理旧构建产物
cleanup() {
    echo_info "Cleaning up old build artifacts..."
    rm -rf ./releases
    rm -f ./tracely ./tracely.exe ./hashpwd ./hashpwd.exe
}

# 构建前端 Dashboard
build_frontend() {
    echo_info "Building frontend dashboard..."
    if [ ! -d "dashboard/node_modules" ]; then
        echo_info "Installing dashboard dependencies..."
        cd dashboard
        npm install
        cd ..
    fi
    
    cd dashboard
    npm run build
    cd ..
    
    # 复制前端构建产物到 internal/static
    echo_info "Copying frontend build to internal/static..."
    rm -rf ./internal/static
    cp -r ./dashboard/dist ./internal/static
}

# 编译指定平台的二进制
build_binary() {
    local GOOS=$1
    local GOARCH=$2
    local EXT=""
    
    if [ "$GOOS" = "windows" ]; then
        EXT=".exe"
    fi
    
    local OUTPUT_DIR="./releases/${GOOS}-${GOARCH}"
    local TRACELY_BIN="${PROJECT_NAME}${EXT}"
    local HASHPWD_BIN="hashpwd${EXT}"
    
    echo_info "Building for ${GOOS}/${GOARCH}..."
    
    mkdir -p "$OUTPUT_DIR"
    
    # 编译 tracely（注入版本信息）
    CGO_ENABLED=1 GOOS=$GOOS GOARCH=$GOARCH go build \
        -ldflags="-s -w -X github.com/hanxi/tracely/internal/version.Version=${VERSION} -X github.com/hanxi/tracely/internal/version.BuildTime=${BUILD_TIME} -X github.com/hanxi/tracely/internal/version.GitCommit=${GIT_COMMIT} -X github.com/hanxi/tracely/internal/version.GoVersion=$(go version | cut -d' ' -f3)" \
        -o "${OUTPUT_DIR}/${TRACELY_BIN}" \
        .
    
    # 编译 hashpwd
    CGO_ENABLED=1 GOOS=$GOOS GOARCH=$GOARCH go build \
        -ldflags="-s -w" \
        -o "${OUTPUT_DIR}/${HASHPWD_BIN}" \
        ./cmd/hashpwd
    
    # 复制配置文件和说明
    cp ./config.example.yaml "${OUTPUT_DIR}/"
    cp ./README.md "${OUTPUT_DIR}/"
    
    echo_info "Built successfully: ${OUTPUT_DIR}/"
}

# 打包为 zip
package_zip() {
    local OUTPUT_DIR=$1
    local PLATFORM=$2
    local ZIP_FILE="./releases/${PROJECT_NAME}-${VERSION}-${PLATFORM}.zip"
    
    echo_info "Packaging ${PLATFORM}..."
    
    cd "$OUTPUT_DIR"
    zip -r "../../${ZIP_FILE}" ./*
    cd ../..
    
    echo_info "Created: ${ZIP_FILE}"
}

# 打包为 tar.gz
package_tar() {
    local OUTPUT_DIR=$1
    local PLATFORM=$2
    local TAR_FILE="./releases/${PROJECT_NAME}-${VERSION}-${PLATFORM}.tar.gz"
    
    echo_info "Packaging ${PLATFORM}..."
    
    cd "$OUTPUT_DIR"
    tar -czf "../../${ZIP_FILE}" ./*
    cd ../..
    
    echo_info "Created: ${TAR_FILE}"
}

# 主函数
main() {
    echo_info "========================================="
    echo_info "  ${PROJECT_NAME} Release Builder"
    echo_info "  Version: ${VERSION}"
    echo_info "========================================="
    
    # 清理
    cleanup
    
    # 构建前端
    build_frontend
    
    # 定义目标平台
    declare -a PLATFORMS=(
        "linux/amd64"
        "linux/arm64"
        "darwin/amd64"
        "darwin/arm64"
        "windows/amd64"
        "windows/arm64"
    )
    
    # 编译各平台二进制
    for PLATFORM in "${PLATFORMS[@]}"; do
        IFS='/' read -r -a ARRAY <<< "$PLATFORM"
        GOOS=${ARRAY[0]}
        GOARCH=${ARRAY[1]}
        build_binary $GOOS $GOARCH
    done
    
    # 打包
    echo_info "========================================="
    echo_info "  Packaging releases..."
    echo_info "========================================="
    
    for PLATFORM in "${PLATFORMS[@]}"; do
        IFS='/' read -r -a ARRAY <<< "$PLATFORM"
        GOOS=${ARRAY[0]}
        GOARCH=${ARRAY[1]}
        OUTPUT_DIR="./releases/${GOOS}-${GOARCH}"
        
        if [ "$GOOS" = "windows" ]; then
            package_zip "$OUTPUT_DIR" "${GOOS}-${GOARCH}"
        else
            cd "$OUTPUT_DIR"
            tar -czf "../../${PROJECT_NAME}-${VERSION}-${GOOS}-${GOARCH}.tar.gz" ./*
            cd ../..
            echo_info "Created: ./releases/${PROJECT_NAME}-${VERSION}-${GOOS}-${GOARCH}.tar.gz"
        fi
    done
    
    # 生成 checksums
    echo_info "========================================="
    echo_info "  Generating checksums..."
    echo_info "========================================="
    
    cd ./releases
    sha256sum ${PROJECT_NAME}-* > SHA256SUMS.txt
    cd ..
    
    echo_info "Created: ./releases/SHA256SUMS.txt"
    
    echo_info "========================================="
    echo_info "  Build completed successfully!"
    echo_info "========================================="
    echo_info ""
    echo_info "Release files are in ./releases/"
    ls -lh ./releases/
}

# 执行
main "$@"

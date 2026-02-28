#!/bin/bash
# Tracely 一键配置生成脚本
# 使用方法：./scripts/gen-config.sh

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
CONFIG_DIR="$PROJECT_ROOT/config"
CONFIG_FILE="$CONFIG_DIR/config.yaml"
CONFIG_EXAMPLE="$PROJECT_ROOT/config.example.yaml"

echo "🔧 Tracely 配置生成工具"
echo "========================"
echo ""

# 检查是否存在 config.example.yaml
if [ ! -f "$CONFIG_EXAMPLE" ]; then
    echo "❌ 错误：找不到 config.example.yaml"
    exit 1
fi

# 创建 config 目录（如果不存在）
if [ ! -d "$CONFIG_DIR" ]; then
    mkdir -p "$CONFIG_DIR"
    echo "📁 已创建 config 目录"
fi

# 读取示例配置内容
CONFIG_CONTENT=$(cat "$CONFIG_EXAMPLE")

# 如果 config/config.yaml 已存在，询问是否覆盖
if [ -f "$CONFIG_FILE" ]; then
    echo "⚠️  config/config.yaml 已存在"
    read -p "是否覆盖？(y/N): " confirm
    if [ "$confirm" != "y" ] && [ "$confirm" != "Y" ]; then
        echo "已取消"
        exit 0
    fi
fi

# 写入配置文件（使用 cat 重定向，避免 cp 在挂载卷上的问题）
echo "$CONFIG_CONTENT" > "$CONFIG_FILE"
echo "✅ 已生成 config/config.yaml"
echo ""

# 生成 JWT Secret
echo "🔐 正在生成 JWT Secret..."
JWT_SECRET=$(/app/tracely -generate-secret -secret-length 32 2>/dev/null | grep "Generated secret" -A 1 | tail -n 1 | xargs)
if [ -z "$JWT_SECRET" ]; then
    # 如果命令失败，使用 openssl 生成
    JWT_SECRET=$(openssl rand -hex 16)
fi
echo "   JWT Secret: $JWT_SECRET"

# 生成 App Secret
echo "🔐 正在生成 App Secret..."
APP_SECRET=$(/app/tracely -generate-secret -secret-length 32 2>/dev/null | grep "Generated secret" -A 1 | tail -n 1 | xargs)
if [ -z "$APP_SECRET" ]; then
    APP_SECRET=$(openssl rand -hex 16)
fi
echo "   App Secret: $APP_SECRET"

# 询问用户密码
echo ""
echo "👤 设置管理员密码"
read -p "请输入 admin 用户的密码（默认：admin123）: " password
if [ -z "$password" ]; then
    password="admin123"
fi

# 生成密码哈希
echo "🔐 正在生成密码哈希..."
PASSWORD_HASH=$(/app/tracely -hashpwd -password "$password" 2>/dev/null | grep "Password hash" -A 1 | tail -n 1 | xargs)
if [ -z "$PASSWORD_HASH" ]; then
    echo "❌ 生成密码哈希失败"
    exit 1
fi
echo "   密码哈希已生成"

# 更新配置文件
echo ""
echo "📝 正在更新配置文件..."

# 使用 sed 替换配置（输出到临时变量，避免在挂载卷上创建临时文件）
UPDATED_CONTENT=$(echo "$CONFIG_CONTENT" | \
    sed "s|your-jwt-secret-please-change-this-to-32-chars|$JWT_SECRET|g" | \
    sed "s|my-app-secret-please-change-this-to-32-chars|$APP_SECRET|g" | \
    sed "s|\$2a\$10\$xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx|$PASSWORD_HASH|g")

# 一次性写入更新后的内容
echo "$UPDATED_CONTENT" > "$CONFIG_FILE"

echo "✅ 配置生成完成！"
echo ""
echo "📋 配置摘要:"
echo "   - JWT Secret: 已生成"
echo "   - App Secret: 已生成"
echo "   - 管理员密码：$password"
echo ""
echo "🚀 下一步:"
echo "   1. 检查 config/config.yaml 配置是否正确"
echo "   2. 运行：docker compose up -d"
echo "   3. 访问：http://localhost:3001"
echo "   4. 使用 admin / $password 登录 Dashboard"
echo ""

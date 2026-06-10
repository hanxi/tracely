#!/bin/bash
# Tracely 配置管理工具
# 使用方法：
#   ./scripts/gen-config.sh              # 交互式菜单
#   ./scripts/gen-config.sh init         # 初始化配置
#   ./scripts/gen-config.sh show         # 查看配置摘要
#   ./scripts/gen-config.sh help         # 查看帮助

set -e

# ============================================================
# 常量和路径
# ============================================================

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
CONFIG_DIR="$PROJECT_ROOT/config"
CONFIG_FILE="$CONFIG_DIR/config.yaml"
CONFIG_EXAMPLE="$PROJECT_ROOT/config.example.yaml"

# ============================================================
# 工具函数
# ============================================================

detect_tracely_binary() {
    if [ -n "$TRACELY_BIN" ] && [ -x "$TRACELY_BIN" ]; then
        echo "$TRACELY_BIN"
    elif [ -x "/app/tracely" ]; then
        echo "/app/tracely"
    elif [ -x "$PROJECT_ROOT/tracely" ]; then
        echo "$PROJECT_ROOT/tracely"
    elif [ -x "./tracely" ]; then
        echo "./tracely"
    else
        echo ""
    fi
}

TRACELY_BIN="$(detect_tracely_binary)"

generate_secret() {
    local length="${1:-32}"
    local secret=""
    if [ -n "$TRACELY_BIN" ]; then
        secret=$("$TRACELY_BIN" -generate-secret -secret-length "$length" 2>/dev/null | tail -n 1 | xargs)
    fi
    if [ -z "$secret" ]; then
        secret=$(openssl rand -hex "$length" 2>/dev/null) || secret=$(head -c "$length" /dev/urandom | od -An -tx1 | tr -d ' \n')
    fi
    echo "$secret"
}

hash_password() {
    local pw="$1"
    local hash=""
    if [ -n "$TRACELY_BIN" ]; then
        hash=$("$TRACELY_BIN" -hashpwd -password "$pw" 2>/dev/null | grep '^\$2')
    fi
    if [ -z "$hash" ]; then
        echo "错误：未找到 tracely 二进制文件，密码哈希需要 tracely 二进制。" >&2
        echo "请先构建：go build -o tracely ." >&2
        return 1
    fi
    echo "$hash"
}

mask_secret() {
    local s="$1"
    local len=${#s}
    if [ "$len" -le 8 ]; then
        echo "****"
    else
        echo "${s:0:4}****${s:$((len-4)):4}"
    fi
}

ensure_config_exists() {
    if [ ! -f "$CONFIG_FILE" ]; then
        echo "错误：配置文件不存在: $CONFIG_FILE"
        echo "请先运行: $0 init"
        exit 1
    fi
}

# ============================================================
# YAML 读写函数
# ============================================================

yaml_get_jwt_secret() {
    awk '/^jwt:/{found=1; next} found && /^[^ ]/{found=0} found && /secret:/{gsub(/.*secret: *"?/, ""); gsub(/".*/, ""); print; exit}' "$CONFIG_FILE"
}

yaml_set_jwt_secret() {
    local new_secret="$1"
    local content
    content=$(awk -v val="$new_secret" '
        /^jwt:/ { in_jwt=1 }
        in_jwt && /^[^ ]/ && !/^jwt:/ { in_jwt=0 }
        in_jwt && /secret:/ {
            sub(/secret: *"[^"]*"/, "secret: \"" val "\"")
            sub(/secret: *[^"[:space:]][^[:space:]]*/, "secret: \"" val "\"")
        }
        { print }
    ' "$CONFIG_FILE")
    echo "$content" > "$CONFIG_FILE"
}

yaml_list_apps() {
    awk '
        /^apps:/ { in_apps=1; next }
        in_apps && /^[^ #]/ { exit }
        in_apps && /- appId:/ {
            id=$0; gsub(/.*appId: *"?/, "", id); gsub(/".*/, "", id)
        }
        in_apps && /appName:/ {
            name=$0; gsub(/.*appName: *"?/, "", name); gsub(/".*/, "", name)
        }
        in_apps && /appSecret:/ {
            secret=$0; gsub(/.*appSecret: *"?/, "", secret); gsub(/".*/, "", secret)
            printf "%s\t%s\t%s\n", id, name, secret
        }
    ' "$CONFIG_FILE"
}

yaml_set_app_field() {
    local target_id="$1"
    local field="$2"
    local new_value="$3"
    local content
    content=$(awk -v tid="$target_id" -v fld="$field" -v val="$new_value" '
        /^apps:/ { in_apps=1 }
        in_apps && /^[^ #]/ && !/^apps:/ { in_apps=0 }
        in_apps && /- appId:/ {
            cur=$0; gsub(/.*appId: *"?/, "", cur); gsub(/".*/, "", cur)
            matched=(cur == tid)
        }
        matched && $0 ~ fld ":" {
            if (!sub(fld ": *\"[^\"]*\"", fld ": \"" val "\"")) {
                sub(fld ": *[^[:space:]]+", fld ": \"" val "\"")
            }
            print
            matched=0
            next
        }
        { print }
    ' "$CONFIG_FILE")
    echo "$content" > "$CONFIG_FILE"
}

yaml_add_app() {
    local app_id="$1"
    local app_name="$2"
    local app_secret="$3"
    local content
    content=$(awk -v id="$app_id" -v name="$app_name" -v secret="$app_secret" '
        /^apps:/ { in_apps=1 }
        in_apps && /appSecret:/ { last_secret=NR }
        in_apps && /^[^ #]/ && !/^apps:/ { in_apps=0 }
        { lines[NR]=$0 }
        END {
            for (i=1; i<=NR; i++) {
                print lines[i]
                if (i == last_secret) {
                    printf "  - appId: \"%s\"\n    appName: \"%s\"\n    appSecret: \"%s\"\n", id, name, secret
                }
            }
            if (!last_secret) {
                printf "  - appId: \"%s\"\n    appName: \"%s\"\n    appSecret: \"%s\"\n", id, name, secret
            }
        }
    ' "$CONFIG_FILE")
    echo "$content" > "$CONFIG_FILE"
}

yaml_remove_app() {
    local target_id="$1"
    local content
    content=$(awk -v tid="$target_id" '
        /^apps:/ { in_apps=1 }
        in_apps && /^[^ #]/ && !/^apps:/ { in_apps=0 }
        in_apps && /- appId:/ {
            cur=$0; gsub(/.*appId: *"?/, "", cur); gsub(/".*/, "", cur)
            if (cur == tid) { skip=1; next }
        }
        skip && /^    [^ -]/ { next }
        skip { skip=0 }
        { print }
    ' "$CONFIG_FILE")
    echo "$content" > "$CONFIG_FILE"
}

yaml_list_users() {
    awk '
        /^users:/ { in_users=1; next }
        in_users && /^[^ #]/ { exit }
        in_users && /- username:/ {
            name=$0; gsub(/.*username: *"?/, "", name); gsub(/".*/, "", name)
            print name
        }
    ' "$CONFIG_FILE"
}

yaml_set_user_password() {
    local target_user="$1"
    local new_hash="$2"
    local content
    content=$(awk -v tuser="$target_user" -v nhash="$new_hash" '
        /^users:/ { in_users=1 }
        in_users && /^[^ #]/ && !/^users:/ { in_users=0 }
        in_users && /- username:/ {
            cur=$0; gsub(/.*username: *"?/, "", cur); gsub(/".*/, "", cur)
            matched=(cur == tuser)
        }
        matched && /passwordHash:/ {
            indent=$0; sub(/[^ ].*/, "", indent)
            print indent "passwordHash: \"" nhash "\""
            matched=0
            next
        }
        { print }
    ' "$CONFIG_FILE")
    echo "$content" > "$CONFIG_FILE"
}

yaml_app_exists() {
    local target_id="$1"
    yaml_list_apps | awk -F'\t' -v tid="$target_id" '$1 == tid { found=1 } END { exit !found }'
}

# ============================================================
# 子命令实现
# ============================================================

cmd_init() {
    echo "Tracely 配置初始化"
    echo "========================"
    echo ""

    if [ ! -f "$CONFIG_EXAMPLE" ]; then
        echo "错误：找不到 config.example.yaml"
        exit 1
    fi

    if [ ! -d "$CONFIG_DIR" ]; then
        mkdir -p "$CONFIG_DIR"
        echo "已创建 config 目录"
    fi

    if [ -f "$CONFIG_FILE" ]; then
        echo "config/config.yaml 已存在"
        read -p "是否覆盖？(y/N): " confirm
        if [ "$confirm" != "y" ] && [ "$confirm" != "Y" ]; then
            echo "已取消"
            return
        fi
    fi

    CONFIG_CONTENT=$(cat "$CONFIG_EXAMPLE")

    echo "正在生成 JWT Secret..."
    JWT_SECRET=$(generate_secret 32)
    echo "  JWT Secret: $(mask_secret "$JWT_SECRET")"

    echo "正在生成 App Secret..."
    APP_SECRET=$(generate_secret 32)
    echo "  App Secret: $(mask_secret "$APP_SECRET")"

    echo ""
    read -p "请输入 admin 用户的密码（默认：admin123）: " password
    if [ -z "$password" ]; then
        password="admin123"
    fi

    echo "正在生成密码哈希..."
    PASSWORD_HASH=$(hash_password "$password")
    if [ $? -ne 0 ] || [ -z "$PASSWORD_HASH" ]; then
        echo "生成密码哈希失败"
        exit 1
    fi

    UPDATED_CONTENT=$(echo "$CONFIG_CONTENT" | \
        sed "s|your-jwt-secret-please-change-this-to-32-chars|$JWT_SECRET|g" | \
        sed "s|my-app-secret-please-change-this-to-32-chars|$APP_SECRET|g")

    UPDATED_CONTENT=$(echo "$UPDATED_CONTENT" | awk -v hash="$PASSWORD_HASH" '{
        if ($0 ~ /passwordHash:/) {
            indent=$0; sub(/[^ ].*/, "", indent)
            print indent "passwordHash: \"" hash "\""
        } else {
            print
        }
    }')

    echo "$UPDATED_CONTENT" > "$CONFIG_FILE"

    echo ""
    echo "配置生成完成！"
    echo ""
    echo "配置摘要:"
    echo "  - JWT Secret: 已生成"
    echo "  - App Secret: 已生成"
    echo "  - 管理员密码: $password"
    echo ""
    echo "下一步:"
    echo "  1. 检查 config/config.yaml 配置是否正确"
    echo "  2. 运行：docker compose up -d"
    echo "  3. 访问：http://localhost:3001"
    echo "  4. 使用 admin / $password 登录 Dashboard"
}

cmd_show() {
    ensure_config_exists
    echo "Tracely 配置摘要"
    echo "========================"

    local port jwt_secret
    port=$(awk '/^port:/{gsub(/.*: *"?/, ""); gsub(/".*/, ""); print; exit}' "$CONFIG_FILE")
    jwt_secret=$(yaml_get_jwt_secret)

    echo "端口: ${port:-未设置}"
    echo "JWT Secret: $(mask_secret "${jwt_secret:-未设置}")"
    echo ""

    echo "应用列表:"
    local app_count=0
    while IFS=$'\t' read -r id name secret; do
        app_count=$((app_count + 1))
        echo "  [$app_count] $id ($name)  Secret: $(mask_secret "$secret")"
    done < <(yaml_list_apps)
    if [ "$app_count" -eq 0 ]; then
        echo "  (无)"
    fi
    echo ""

    echo "用户列表:"
    local user_count=0
    while IFS= read -r username; do
        user_count=$((user_count + 1))
        echo "  [$user_count] $username"
    done < <(yaml_list_users)
    if [ "$user_count" -eq 0 ]; then
        echo "  (无)"
    fi
}

cmd_set_jwt_secret() {
    ensure_config_exists
    local old_secret new_secret
    old_secret=$(yaml_get_jwt_secret)
    echo "当前 JWT Secret: $(mask_secret "$old_secret")"
    echo ""

    new_secret=$(generate_secret 32)
    yaml_set_jwt_secret "$new_secret"

    echo "JWT Secret 已更新: $(mask_secret "$new_secret")"
    echo "注意：更新后需要重启服务，且所有已登录用户需重新登录。"
}

cmd_set_password() {
    ensure_config_exists
    local users=()
    while IFS= read -r u; do
        users+=("$u")
    done < <(yaml_list_users)

    if [ ${#users[@]} -eq 0 ]; then
        echo "错误：配置中没有用户"
        return 1
    fi

    local target_user
    if [ ${#users[@]} -eq 1 ]; then
        target_user="${users[0]}"
        echo "用户: $target_user"
    else
        echo "请选择要修改密码的用户:"
        for i in "${!users[@]}"; do
            echo "  $((i+1))) ${users[$i]}"
        done
        read -p "请选择 [1-${#users[@]}]: " choice
        if [ -z "$choice" ] || [ "$choice" -lt 1 ] || [ "$choice" -gt ${#users[@]} ] 2>/dev/null; then
            echo "无效选择"
            return 1
        fi
        target_user="${users[$((choice-1))]}"
    fi

    read -p "请输入新密码: " new_password
    if [ -z "$new_password" ]; then
        echo "密码不能为空"
        return 1
    fi

    echo "正在生成密码哈希..."
    local new_hash
    new_hash=$(hash_password "$new_password")
    if [ $? -ne 0 ] || [ -z "$new_hash" ]; then
        echo "生成密码哈希失败"
        return 1
    fi

    yaml_set_user_password "$target_user" "$new_hash"
    echo "用户 $target_user 的密码已更新。"
    echo "注意：更新后需要重启服务。"
}

cmd_app_list() {
    ensure_config_exists
    echo "应用列表"
    echo "========================"
    local count=0
    while IFS=$'\t' read -r id name secret; do
        count=$((count + 1))
        echo "  [$count] AppID:     $id"
        echo "       AppName:   $name"
        echo "       AppSecret: $(mask_secret "$secret")"
        echo ""
    done < <(yaml_list_apps)
    if [ "$count" -eq 0 ]; then
        echo "  (无应用)"
    fi
    echo "共 $count 个应用"
}

cmd_app_add() {
    ensure_config_exists
    local app_id="${1:-}"
    local app_name="${2:-}"

    if [ -z "$app_id" ]; then
        read -p "请输入 App ID: " app_id
    fi
    if [ -z "$app_id" ]; then
        echo "App ID 不能为空"
        return 1
    fi

    if yaml_app_exists "$app_id"; then
        echo "错误：App ID '$app_id' 已存在"
        return 1
    fi

    if [ -z "$app_name" ]; then
        read -p "请输入应用名称: " app_name
    fi
    if [ -z "$app_name" ]; then
        app_name="$app_id"
    fi

    local app_secret
    app_secret=$(generate_secret 32)

    yaml_add_app "$app_id" "$app_name" "$app_secret"

    echo "应用已添加:"
    echo "  AppID:     $app_id"
    echo "  AppName:   $app_name"
    echo "  AppSecret: $app_secret"
    echo ""
    echo "请妥善保存 AppSecret，后续不会完整显示。"
    echo "注意：更新后需要重启服务。"
}

cmd_app_set_secret() {
    ensure_config_exists
    local target_id="${1:-}"

    if [ -z "$target_id" ]; then
        target_id=$(select_app "请选择要重新生成密钥的应用")
        if [ -z "$target_id" ]; then return 1; fi
    fi

    if ! yaml_app_exists "$target_id"; then
        echo "错误：App ID '$target_id' 不存在"
        return 1
    fi

    local new_secret
    new_secret=$(generate_secret 32)
    yaml_set_app_field "$target_id" "appSecret" "$new_secret"

    echo "应用 '$target_id' 的密钥已更新:"
    echo "  AppSecret: $new_secret"
    echo ""
    echo "请妥善保存新密钥，后续不会完整显示。"
    echo "注意：更新后需要重启服务，并更新 SDK 中的密钥配置。"
}

cmd_app_set_id() {
    ensure_config_exists
    local old_id="${1:-}"
    local new_id="${2:-}"

    if [ -z "$old_id" ]; then
        old_id=$(select_app "请选择要修改 ID 的应用")
        if [ -z "$old_id" ]; then return 1; fi
    fi

    if ! yaml_app_exists "$old_id"; then
        echo "错误：App ID '$old_id' 不存在"
        return 1
    fi

    if [ -z "$new_id" ]; then
        read -p "请输入新的 App ID: " new_id
    fi
    if [ -z "$new_id" ]; then
        echo "App ID 不能为空"
        return 1
    fi

    if [ "$old_id" = "$new_id" ]; then
        echo "新旧 App ID 相同，无需修改"
        return 0
    fi

    if yaml_app_exists "$new_id"; then
        echo "错误：App ID '$new_id' 已存在"
        return 1
    fi

    yaml_set_app_field "$old_id" "appId" "$new_id"

    echo "App ID 已更新: $old_id -> $new_id"
    echo "注意：更新后需要重启服务，并更新 SDK 中的 appId 配置。"
}

cmd_app_remove() {
    ensure_config_exists
    local target_id="${1:-}"

    if [ -z "$target_id" ]; then
        target_id=$(select_app "请选择要删除的应用")
        if [ -z "$target_id" ]; then return 1; fi
    fi

    if ! yaml_app_exists "$target_id"; then
        echo "错误：App ID '$target_id' 不存在"
        return 1
    fi

    read -p "确认删除应用 '$target_id'？(y/N): " confirm
    if [ "$confirm" != "y" ] && [ "$confirm" != "Y" ]; then
        echo "已取消"
        return 0
    fi

    yaml_remove_app "$target_id"
    echo "应用 '$target_id' 已删除。"
    echo "注意：更新后需要重启服务。"
}

select_app() {
    local prompt="${1:-请选择应用}"
    local apps=()
    local ids=()
    while IFS=$'\t' read -r id name secret; do
        apps+=("$id ($name)")
        ids+=("$id")
    done < <(yaml_list_apps)

    if [ ${#apps[@]} -eq 0 ]; then
        echo "错误：没有已配置的应用" >&2
        return 1
    fi

    echo "$prompt:" >&2
    for i in "${!apps[@]}"; do
        echo "  $((i+1))) ${apps[$i]}" >&2
    done
    read -p "请选择 [1-${#apps[@]}]: " choice <&0
    if [ -z "$choice" ] || [ "$choice" -lt 1 ] || [ "$choice" -gt ${#apps[@]} ] 2>/dev/null; then
        echo "无效选择" >&2
        return 1
    fi
    echo "${ids[$((choice-1))]}"
}

# ============================================================
# 交互式菜单
# ============================================================

menu_app() {
    while true; do
        echo ""
        echo "应用管理"
        echo "========================"
        echo "  1) 查看应用列表"
        echo "  2) 添加新应用"
        echo "  3) 重新生成应用密钥"
        echo "  4) 修改应用 ID"
        echo "  5) 删除应用"
        echo "  0) 返回"
        echo ""
        read -p "请选择 [0-5]: " choice
        echo ""
        case "$choice" in
            1) cmd_app_list ;;
            2) cmd_app_add ;;
            3) cmd_app_set_secret ;;
            4) cmd_app_set_id ;;
            5) cmd_app_remove ;;
            0) return ;;
            *) echo "无效选择" ;;
        esac
    done
}

cmd_menu() {
    echo "Tracely 配置管理工具"
    echo "========================"

    while true; do
        echo ""
        echo "请选择操作:"
        echo "  1) 初始化新配置"
        echo "  2) 查看当前配置"
        echo "  3) 重新生成 JWT Secret"
        echo "  4) 修改用户密码"
        echo "  5) 管理应用 -->"
        echo "  0) 退出"
        echo ""
        read -p "请选择 [0-5]: " choice
        echo ""
        case "$choice" in
            1) cmd_init ;;
            2) cmd_show ;;
            3) cmd_set_jwt_secret ;;
            4) cmd_set_password ;;
            5) menu_app ;;
            0) echo "再见！"; exit 0 ;;
            *) echo "无效选择" ;;
        esac
    done
}

# ============================================================
# 帮助信息
# ============================================================

show_help() {
    echo "Tracely 配置管理工具"
    echo ""
    echo "用法: $0 [命令] [参数]"
    echo ""
    echo "命令:"
    echo "  (无)                           交互式菜单"
    echo "  init                           初始化配置（从模板生成）"
    echo "  show                           查看当前配置摘要"
    echo "  set-jwt-secret                 重新生成 JWT Secret"
    echo "  set-password                   修改用户密码"
    echo "  app list                       查看应用列表"
    echo "  app add [appId] [appName]      添加新应用"
    echo "  app set-secret [appId]         重新生成应用密钥"
    echo "  app set-id [oldId] [newId]     修改应用 ID"
    echo "  app remove [appId]             删除应用"
    echo "  help                           显示此帮助信息"
}

# ============================================================
# 入口分发
# ============================================================

case "${1:-}" in
    init)             cmd_init ;;
    show)             cmd_show ;;
    set-jwt-secret)   cmd_set_jwt_secret ;;
    set-password)     cmd_set_password ;;
    app)
        case "${2:-}" in
            list)       cmd_app_list ;;
            add)        cmd_app_add "${3:-}" "${4:-}" ;;
            set-secret) cmd_app_set_secret "${3:-}" ;;
            set-id)     cmd_app_set_id "${3:-}" "${4:-}" ;;
            remove)     cmd_app_remove "${3:-}" ;;
            *)          show_help; exit 1 ;;
        esac
        ;;
    -h|--help|help) show_help ;;
    "")             cmd_menu ;;
    *)              echo "未知命令: $1"; echo ""; show_help; exit 1 ;;
esac

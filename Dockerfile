# 多阶段构建 Dockerfile
# 阶段 1：构建前端
FROM oven/bun:1-alpine AS frontend-builder

WORKDIR /app/dashboard

COPY dashboard/package*.json ./
RUN bun install

COPY dashboard/ ./
RUN bun run build

# 阶段 2：编译 Go 后端
FROM golang:1.26-alpine AS backend-builder

WORKDIR /app

# 安装构建依赖
RUN apk add --no-cache git gcc musl-dev

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# 复制已构建的前端资源
COPY --from=frontend-builder /app/dashboard/dist ./dashboard/dist

# 编译二进制
RUN CGO_ENABLED=1 GOOS=linux go build -o tracely .

# 阶段 3：运行镜像
FROM alpine:latest

# 版本信息参数（通过 --build-arg 传入）
ARG VERSION=unknown
ARG GIT_COMMIT=unknown
ARG BUILD_TIME=unknown

WORKDIR /app

# 安装运行时依赖（SQLite 需要）
RUN apk add --no-cache ca-certificates

# 复制二进制文件
COPY --from=backend-builder /app/tracely .

# 创建数据目录
RUN mkdir -p /app/data

# 挂载数据目录
VOLUME ["/app/data"]

# 暴露端口
EXPOSE 3001

# 设置 OCI 标准镜像元数据
LABEL org.opencontainers.image.version="${VERSION}" \
      org.opencontainers.image.revision="${GIT_COMMIT}" \
      org.opencontainers.image.created="${BUILD_TIME}" \
      org.opencontainers.image.title="Tracely" \
      org.opencontainers.image.description="Lightweight Error Tracking System" \
      org.opencontainers.image.source="https://github.com/hanxi/tracely"

# 启动服务
CMD ["./tracely"]
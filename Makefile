.PHONY: build build-frontend build-backend dev docker docker-push clean

# 一键构建全部
build: build-frontend build-backend

# 构建前端 Dashboard
build-frontend:
	cd dashboard && bun install && bun run build

# 构建后端（嵌入前端资源）
build-backend: build-frontend
	go build -o tracely .

# 开发模式
dev:
	go run .

# 构建 Docker 镜像
docker:
	docker build -t hanxi/tracely:latest .

# 推送 Docker 镜像到 Docker Hub
docker-push:
	@echo "Pushing Docker image to Docker Hub..."
	docker push hanxi/tracely:latest

# 清理构建产物
clean:
	rm -rf dist/ build/ tracely tracely.exe
	rm -rf dashboard/dist
	rm -rf sdk/ts/dist

.PHONY: build build-frontend build-backend dev docker docker-push clean release release-publish release-all release-auto changelog version

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
	rm -rf releases/
	rm -rf internal/static

# 编译多平台发布包
release:
	@echo "Building release packages..."
	@if [ -z "$(VERSION)" ]; then \
		echo "Error: VERSION is required. Usage: make release VERSION=1.0.0"; \
		exit 1; \
	fi
	bash ./scripts/release.sh

# 发布到 GitHub Release
release-publish:
	@echo "Publishing to GitHub Release..."
	@if [ -z "$(VERSION)" ]; then \
		echo "Error: VERSION is required. Usage: make release-publish VERSION=1.0.0"; \
		exit 1; \
	fi
	bash ./scripts/publish-release.sh

# 完整发布流程（编译 + 发布）
release-all: release release-publish
	@echo "Release completed!"

# 自动发布（打 tag 并推送）
release-auto:
	@echo "Running auto-release script..."
	bash ./scripts/auto-release.sh $(VERSION)

# 生成 CHANGELOG
changelog:
	@echo "Generating CHANGELOG..."
	@if ! command -v git-chglog &> /dev/null; then \
		echo "Installing git-chglog..."; \
		go install github.com/git-chglog/git-chglog/cmd/git-chglog@latest; \
	fi
	git-chglog -o CHANGELOG.md

# 显示当前版本号
version:
	@go run -ldflags="-X github.com/hanxi/tracely/internal/version.Version=dev" . -version

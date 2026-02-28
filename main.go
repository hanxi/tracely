package main

import (
	"crypto/rand"
	"flag"
	"fmt"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/hanxi/tracely/dashboard"
	"github.com/hanxi/tracely/internal/config"
	"github.com/hanxi/tracely/internal/handler"
	"github.com/hanxi/tracely/internal/middleware"
	"github.com/hanxi/tracely/internal/model"
	"github.com/hanxi/tracely/internal/version"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	// 定义命令行参数
	serverMode := flag.Bool("server", false, "启动服务器模式")
	hashpwdMode := flag.Bool("hashpwd", false, "密码哈希模式")
	password := flag.String("password", "", "要哈希的密码（与 -hashpwd 一起使用）")
	generateSecret := flag.Bool("generate-secret", false, "生成随机 Secret")
	secretLength := flag.Int("secret-length", 32, "生成 Secret 的长度（与 -generate-secret 一起使用）")
	showVersion := flag.Bool("version", false, "显示版本信息")
	showVersionShort := flag.Bool("v", false, "显示版本信息（简写）")

	flag.Parse()

	// 显示版本信息
	if *showVersion || *showVersionShort {
		printVersion()
		return
	}

	// 密码哈希模式
	if *hashpwdMode {
		if *password == "" {
			fmt.Println("Usage: ./tracely -hashpwd -password <yourpassword>")
			os.Exit(1)
		}
		hash, err := bcrypt.GenerateFromPassword([]byte(*password), bcrypt.DefaultCost)
		if err != nil {
			fmt.Printf("Error generating hash: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Password hash for '%s':\n%s\n", *password, string(hash))
		fmt.Println("\nCopy this hash to config.yaml users[].passwordHash field")
		return
	}

	// 生成随机 Secret 模式
	if *generateSecret {
		secret, err := generateSecureRandom(*secretLength)
		if err != nil {
			fmt.Printf("Error generating secret: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Generated secret (%d bytes):\n%s\n", *secretLength, secret)
		return
	}

	// 服务器模式（默认）
	if !*serverMode && !*hashpwdMode && !*generateSecret {
		// 默认启动服务器
		runServer()
		return
	}

	// 如果指定了 server 模式
	if *serverMode {
		runServer()
		return
	}

	// 显示帮助信息
	fmt.Println("Tracely - 轻量级错误追踪系统")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  ./tracely                    # 启动服务器（默认）")
	fmt.Println("  ./tracely -server            # 启动服务器")
	fmt.Println("  ./tracely -version           # 显示版本信息")
	fmt.Println("  ./tracely -hashpwd -password <password>  # 生成密码哈希")
	fmt.Println("  ./tracely -generate-secret [-secret-length 32]  # 生成随机 Secret")
	fmt.Println()
	fmt.Println("Options:")
	flag.PrintDefaults()
}

// printVersion 打印版本信息
func printVersion() {
	fmt.Printf("Tracely %s\n", version.Version)
	fmt.Printf("Build time: %s\n", version.BuildTime)
	fmt.Printf("Git commit: %s\n", version.GitCommit)
	fmt.Printf("Go version: %s\n", version.GoVersion)
}

func runServer() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	logger.Info("[Tracely] Starting server...")

	// 1. 加载配置
	cfg, err := config.Load()
	if err != nil {
		logger.Error("[Tracely] Failed to load config", "error", err)
		os.Exit(1)
	}

	// 2. 初始化数据库
	db, err := model.InitDB(cfg.DBPath)
	if err != nil {
		logger.Error("[Tracely] Failed to initialize database", "error", err)
		os.Exit(1)
	}

	// 3. 启动 Nonce 清理任务
	middleware.StartNonceCleaner(cfg.NonceTTL)

	// 4. 启动活跃日志定时清理任务
	model.StartActiveLogCleanup(db, cfg.ActiveLogRetentionDays)

	// 5. 创建 Gin 实例
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	// 跨域中间件
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization, X-App-Id, X-Timestamp, X-Nonce, X-Signature")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	})

	// 6. 注册路由
	// 登录接口（无需认证）
	r.POST("/auth/login", handler.Login(cfg))

	// API 接口组（JWT 验证，Dashboard 调用）
	api := r.Group("/api")
	api.Use(middleware.JWTAuth(cfg.JWT.Secret))
	{
		api.GET("/apps", handler.GetApps(cfg))     // 应用列表
		api.GET("/overview", handler.Overview(db)) // 概览数据
		api.GET("/errors", handler.ErrorList(db))  // 错误列表
		api.GET("/stats", handler.Stats(db))       // 活跃统计
	}

	// 上报接口组（HMAC 签名验证 + 限速，SDK 调用）
	report := r.Group("/report")
	report.Use(middleware.RateLimit(cfg.RateLimit))
	report.Use(middleware.SignAuth(cfg))
	{
		report.POST("/error", handler.ReportError(db))
		report.POST("/active", handler.ReportActive(db))
	}

	// 7. 配置静态文件服务（内嵌前端资源）
	staticFS, err := fs.Sub(dashboard.Dist, "dist")
	if err != nil {
		logger.Error("[Tracely] Failed to load static files", "error", err)
		os.Exit(1)
	}
	r.StaticFS("/static", http.FS(staticFS))

	// 8. 前端路由处理：根路径直接返回 index.html
	r.GET("/", func(c *gin.Context) {
		data, err := fs.ReadFile(staticFS, "index.html")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load page"})
			return
		}
		c.Data(http.StatusOK, "text/html; charset=utf-8", data)
	})

	// 9. SPA 路由支持：所有未匹配的路由返回 index.html
	r.NoRoute(func(c *gin.Context) {
		// API 路由不处理
		if strings.HasPrefix(c.Request.URL.Path, "/api/") ||
			strings.HasPrefix(c.Request.URL.Path, "/auth/") ||
			strings.HasPrefix(c.Request.URL.Path, "/report/") ||
			strings.HasPrefix(c.Request.URL.Path, "/static/") {
			c.JSON(http.StatusNotFound, gin.H{"error": "Not Found"})
			return
		}

		// 返回 index.html
		data, err := fs.ReadFile(staticFS, "index.html")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load page"})
			return
		}
		c.Data(http.StatusOK, "text/html; charset=utf-8", data)
	})

	// 10. 启动服务
	addr := fmt.Sprintf(":%s", cfg.Port)
	logger.Info("[Tracely] Server started on port", "port", cfg.Port)
	if err := r.Run(addr); err != nil {
		logger.Error("[Tracely] Failed to start server", "error", err)
		os.Exit(1)
	}
}

// generateSecureRandom 生成安全随机字符串（用于生成 Secret）
func generateSecureRandom(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", bytes), nil
}

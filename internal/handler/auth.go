package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/hanxi/tracely/internal/config"
	"github.com/hanxi/tracely/internal/middleware"
)

// LoginRequest 登录请求
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// Login 登录接口
func Login(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req LoginRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
			return
		}

		// 查找用户
		user, ok := cfg.GetUser(req.Username)
		if !ok {
			// 故意不区分"用户不存在"和"密码错误"，防止用户名枚举
			c.JSON(http.StatusUnauthorized, gin.H{"error": "用户名或密码错误"})
			return
		}

		// 验证密码
		if !user.VerifyPassword(req.Password) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "用户名或密码错误"})
			return
		}

		// 生成 JWT Token
		token, err := middleware.GenerateToken(cfg.JWT.Secret, user.Username, cfg.JWT.ExpireHours)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "生成 Token 失败"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"token":    token,
			"username": user.Username,
		})
	}
}

package config

import (
	"fmt"
	"os"
	"sync"

	"github.com/spf13/viper"
	"golang.org/x/crypto/bcrypt"
)

// Config 服务器配置
type Config struct {
	Port         string
	DBPath       string
	RateLimit    int
	NonceTTL     int
	TimestampTTL int
	JWT          JWT
	Apps         []App
	Users        []User
	Events       []EventConfig // 自定义事件配置（白名单）
}

// App 应用配置（SDK 上报用）
type App struct {
	AppID     string `yaml:"appId"`
	AppName   string `yaml:"appName"`
	AppSecret string `yaml:"appSecret"`
}

// User 用户配置（Dashboard 登录用）
type User struct {
	Username     string `yaml:"username"`
	PasswordHash string `yaml:"passwordHash"`
}

// EventConfig 事件配置（白名单）
type EventConfig struct {
	EventName     string `yaml:"eventName"`
	Description   string `yaml:"description"`
	RetentionDays int    `yaml:"retentionDays"` // 数据保留天数（0=永久保留）
}

// JWT JWT 配置
type JWT struct {
	Secret      string `yaml:"secret"`
	ExpireHours int    `yaml:"expireHours"`
}

var (
	configInstance *Config
	configOnce     sync.Once
)

// Load 加载配置（环境变量 > config.yaml > 默认值）
func Load() (*Config, error) {
	var err error
	configOnce.Do(func() {
		configInstance = &Config{
			Port:         "3001",
			DBPath:       "./tracely.db",
			RateLimit:    60,
			NonceTTL:     300,
			TimestampTTL: 300,
			JWT: JWT{
				Secret:      "default-jwt-secret-change-in-production",
				ExpireHours: 24,
			},
		}

		// 尝试读取 config.yaml（支持多个路径）
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
		viper.AddConfigPath(".")
		viper.AddConfigPath("./")
		viper.AddConfigPath("./config")
		viper.AddConfigPath("/app/config")

		if err = viper.ReadInConfig(); err == nil {
			fmt.Printf("[Tracely] Loaded config from: %s\n", viper.ConfigFileUsed())

			// 加载 YAML 配置
			if err = viper.Unmarshal(configInstance); err != nil {
				fmt.Printf("[Tracely] Warning: Failed to unmarshal config: %v\n", err)
				err = nil // 不中断，使用默认值
			}
		} else {
			fmt.Println("[Tracely] No config.yaml found, using environment variables and defaults")
		}

		// 环境变量覆盖（优先级更高）
		if env := os.Getenv("PORT"); env != "" {
			configInstance.Port = env
		}
		if env := os.Getenv("DB_PATH"); env != "" {
			configInstance.DBPath = env
		}
		if env := os.Getenv("RATE_LIMIT"); env != "" {
			fmt.Sscanf(env, "%d", &configInstance.RateLimit)
		}
		if env := os.Getenv("NONCE_TTL"); env != "" {
			fmt.Sscanf(env, "%d", &configInstance.NonceTTL)
		}
		if env := os.Getenv("TIMESTAMP_TTL"); env != "" {
			fmt.Sscanf(env, "%d", &configInstance.TimestampTTL)
		}

		// 验证配置
		if len(configInstance.Apps) == 0 {
			fmt.Println("[Tracely] Warning: No apps configured in config.yaml")
		} else {
			fmt.Printf("[Tracely] Loaded %d apps from config\n", len(configInstance.Apps))
		}

		if len(configInstance.Users) == 0 {
			fmt.Println("[Tracely] Warning: No users configured in config.yaml")
		} else {
			fmt.Printf("[Tracely] Loaded %d users from config\n", len(configInstance.Users))
		}
	})

	return configInstance, err
}

// GetSecret 根据 AppID 获取 Secret
func (c *Config) GetSecret(appID string) (string, bool) {
	for _, app := range c.Apps {
		if app.AppID == appID {
			return app.AppSecret, true
		}
	}
	return "", false
}

// GetUser 根据用户名获取用户
func (c *Config) GetUser(username string) (User, bool) {
	for _, user := range c.Users {
		if user.Username == username {
			return user, true
		}
	}
	return User{}, false
}

// VerifyPassword 验证密码
func (u *User) VerifyPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password))
	return err == nil
}

// IsEventAllowed 检查事件是否在白名单中
func (c *Config) IsEventAllowed(eventName string) bool {
	for _, event := range c.Events {
		if event.EventName == eventName {
			return true
		}
	}
	return false
}

// GetEventConfig 获取事件配置
func (c *Config) GetEventConfig(eventName string) (EventConfig, bool) {
	for _, event := range c.Events {
		if event.EventName == eventName {
			return event, true
		}
	}
	return EventConfig{}, false
}

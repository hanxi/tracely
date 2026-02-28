package model

import (
	"fmt"
	"sync"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	dbInstance *gorm.DB
	dbOnce     sync.Once
)

// InitDB 初始化数据库（SQLite + 性能优化）
func InitDB(path string) (*gorm.DB, error) {
	var err error
	dbOnce.Do(func() {
		dbInstance, err = gorm.Open(sqlite.Open(path), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Info),
		})
		if err != nil {
			err = fmt.Errorf("failed to open database: %w", err)
			return
		}

		// 获取底层 sql.DB
		sqlDB, err := dbInstance.DB()
		if err != nil {
			err = fmt.Errorf("failed to get underlying sql.DB: %w", err)
			return
		}

		// SQLite 性能优化
		// WAL 模式：提升并发写入性能
		sqlDB.Exec("PRAGMA journal_mode=WAL;")
		// 同步模式：NORMAL 在 WAL 模式下足够安全，性能更好
		sqlDB.Exec("PRAGMA synchronous=NORMAL;")
		// 缓存大小：64MB
		sqlDB.Exec("PRAGMA cache_size=-65536;")
		// 临时表存内存
		sqlDB.Exec("PRAGMA temp_store=MEMORY;")

		// 连接池配置（SQLite 只支持单写，避免锁竞争）
		sqlDB.SetMaxOpenConns(1)
		sqlDB.SetMaxIdleConns(1)

		// 自动迁移数据表
		err = dbInstance.AutoMigrate(&ErrorLog{}, &ActiveLog{})
		if err != nil {
			err = fmt.Errorf("failed to auto migrate: %w", err)
			return
		}

		fmt.Printf("[Tracely] Database initialized: %s\n", path)
	})

	return dbInstance, err
}

// GetDB 获取数据库实例
func GetDB() *gorm.DB {
	return dbInstance
}

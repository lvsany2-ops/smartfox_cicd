package main

import (
	"experiment-service/config"
	"experiment-service/routers"
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	routers.RegisterRoutes(r)

	config.InitDB()
	// 允许在 CI/本地集成测试中跳过 OSS 初始化，避免外部依赖导致进程退出
	config.InitOSS()
	// 轻量级存活探针端点，不依赖外部资源
	r.GET("/live", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "alive"})
	})
	r.GET("/health", func(c *gin.Context) {
		// 检查数据库连接
		db := config.DB
		sqlDB, err := db.DB()
		if err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "unhealthy", "error": "database error"})
			return
		}

		if err := sqlDB.Ping(); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "unhealthy", "error": "database ping failed"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"status": "healthy"})
	})
	r.Run(":8082")
}

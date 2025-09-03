package main

import (
	"log"
	"net/http"
	"notification-service/database"
	"notification-service/routers"
	"time"

	"github.com/gin-gonic/gin"
)

func main() {
	// 初始化数据库
	var err error
	maxRetries := 10
	for i := range maxRetries {
		err = database.InitDB()
		if err == nil {
			break
		}
		log.Printf("Database initialization failed (attempt %d/%d): %v", i+1, maxRetries, err)
		time.Sleep(5 * time.Second)
	}

	if err != nil {
		log.Fatalf("Failed to initialize database after %d attempts: %v", maxRetries, err)
	}

	// 初始化路由
	router := routers.InitRouter()

	// 添加健康检查端点
	router.GET("/health", func(c *gin.Context) {
		// 检查数据库连接
		db := database.GetDB()
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

	// 启动服务
	log.Println("Notification service starting on :8083")
	if err := router.Run(":8083"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

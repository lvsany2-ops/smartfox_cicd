package main

import (
	"net/http"
	"submission/core"
	"submission/global"
	"submission/routers"

	"github.com/gin-gonic/gin"
)

func main() {
	// 初始化配置
	core.InitConf()
	// 初始化日志
	global.Log = core.InitLogger()
	//连接数据库
	global.DB = core.InitGorm()
	router := routers.InitRouter()
	router.GET("/health", func(c *gin.Context) {
		// 检查数据库连接
		db := global.DB
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
	router.Run(global.Config.System.Addr()) // listen and serve on
}

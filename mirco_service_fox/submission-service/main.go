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
	router.GET("/live", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "alive"})
	})
	router.Run(global.Config.System.Addr()) // listen and serve on
}

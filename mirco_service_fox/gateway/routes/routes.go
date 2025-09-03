package routes

import (
	"gateway/config"
	"gateway/proxy"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(r *gin.Engine, cfg *config.ServiceConfig) {
	// 健康检查端点
	r.Any("/*path", func(c *gin.Context) {
		// 特殊处理健康检查
		if c.Request.URL.Path == "/health" {
			c.JSON(200, gin.H{"status": "OK"})
			return
		}
		log.Printf("Received request: %s %s", c.Request.Method, c.Request.URL.Path)
		target := proxy.GetTargetService(c.Request.URL.Path, cfg)
		log.Printf("Routing to: %s", target)
		proxy, err := proxy.NewReverseProxy(target)
		if err != nil {
			log.Printf("Failed to create reverse proxy: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create reverse proxy"})
			return
		}
		proxy.ServeHTTP(c.Writer, c.Request)
	})
}

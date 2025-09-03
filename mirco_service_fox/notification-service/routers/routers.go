// routers/router.go
package routers

import (
	"notification-service/controllers"
	"net/http"

	"github.com/gin-gonic/gin"
)

func InitRouter() *gin.Engine {
	router := gin.Default()

	// 通知相关路由
	router.POST("/api/teacher/experiments/notifications", controllers.CreateNotification)
	router.GET("/api/teacher/experiments/notifications", controllers.GetTeacherNotifications)
	router.GET("/api/student/experiments/notifications/:student_id", controllers.GetStudentNotifications)

	// 添加健康检查端点（main.go 中也有，但这里也加一个保险）
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "healthy"})
	})

	return router
}

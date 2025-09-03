// routers/router.go
package routers

import (
	"notification-service/controllers"

	"github.com/gin-gonic/gin"
)

func InitRouter() *gin.Engine {
	router := gin.Default()

	// 通知相关路由
	router.POST("/api/teacher/experiments/notifications", controllers.CreateNotification)
	router.GET("/api/teacher/experiments/notifications", controllers.GetTeacherNotifications)
	router.GET("/api/student/experiments/notifications/:student_id", controllers.GetStudentNotifications)

	// 健康检查端点在 main.go 中实现（带数据库探测），避免重复注册

	return router
}

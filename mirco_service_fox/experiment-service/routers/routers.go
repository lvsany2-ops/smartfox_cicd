package routers

import (
	"experiment-service/controllers"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.Engine) {
	// 保留一个轻量端点作为手工探活使用（集群探针使用 main.go 中定义的 /health 和 /live）
	r.GET("/ping", controllers.HealthCheck)
	student := r.Group("/api/student")
	{

		student.GET("/experiments", controllers.GetExperiments_Student)
		student.GET("/experiments/:experiment_id", controllers.GetExperimentDetail_Student)
	}

	teacher := r.Group("/api/teacher")
	{
		teacher.POST("/experiments", controllers.CreateExperiment)
		teacher.PUT("/experiments/:experiment_id", controllers.UpdateExperiment)
		teacher.DELETE("/experiments/:experiment_id", controllers.DeleteExperiment)
		teacher.POST("/experiments/:experiment_id/uploadFile", controllers.HandleTeacherUpload)
	}

	r.GET("/api/experiments/:experiment_id/files", controllers.HandleStudentListFiles)
	r.GET("/api/experiments/:experiment_id/files/:filename/download", controllers.HandleStudentDownloadFile)

	external := r.Group("/api/experiments")
	{
		external.GET("/getById/:experiment_id", controllers.GetExperimentById)
		external.POST("/validQuestion", controllers.ValidQuestion)
		external.POST("/questionDetail", controllers.GetQuestionDetail)
		external.POST("/experimentDetail", controllers.GetExperimentDetail)
	}
}

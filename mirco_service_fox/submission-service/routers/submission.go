package routers

import (
	"submission/controller"

	"github.com/gin-gonic/gin"
)

func submission(r *gin.Engine) *gin.Engine {
	s := r.Group("/api/student")
	s.POST("/experiments/:experiment_id/save", controller.SaveAnswer)
	s.POST("/experiments/:experiment_id/submit", controller.SubmitExperiment)
	s.GET("/submissions", controller.GetSubmissions)
	s.GET("/submissions/:experiment_id/:student_id/status", controller.GetSubmissionStatus)
	s.POST("/submissions/:submission_id/GetStudentAns", controller.GetStudentAns)
	s.PUT("/submissions/:experiment_id/UpdateExperimentStatus", controller.UpdateExperimentStatusToInProgress)
	s.DELETE("/experiments/:experiment_id/submissions", controller.DeleteExperimentSubmissions)

	return r
}

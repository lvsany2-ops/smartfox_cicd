package routers

import (
	"lh/controller"

	"github.com/gin-gonic/gin"
)

func CollectRoutes(r *gin.Engine) *gin.Engine {
	user := r.Group("/api/auth")
	//注册
	user.POST("/register", controller.Register)
	//登录
	user.POST("/login", controller.Login)
	//返回用户信息
	user.GET("/profile", controller.Info)
	//修改用户信息
	user.PUT("/update", controller.Update)
	r.GET("/api/student_list", controller.GetStudentList)
	TeacherGroup := r.Group("/api/teacher")
	{
		ExperimentRoutes_Teacher(TeacherGroup) // 挂载实验路由
	}
	internal := r.Group("/internal")
	{
		internal.GET("/users/:id", controller.GetUserByID)
	}

	return r

}
func ExperimentRoutes_Teacher(r *gin.RouterGroup) {

	r.GET("/students", controller.GetStudentListWithGroup)
	r.POST("/groups", controller.CreateStudentGroup)
	r.GET("/groups", controller.GetStudentGroup)
	r.PUT("/groups/:group_id", controller.UpdateStudentGroup)
	r.DELETE("/groups/:group_id", controller.DeleteStudentGroup)
}

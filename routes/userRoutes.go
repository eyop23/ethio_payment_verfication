package routes

import (
	"payment_verification/controllers"

	"github.com/gin-gonic/gin"
)

func UserRoutes(r *gin.Engine) {
	u := r.Group("/api/user")
	{
		u.POST("/", controllers.AddUser)
		u.GET("/", controllers.GetUsers)
	}
}

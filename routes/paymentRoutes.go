package routes

import (
	"payment_verification/controllers"

	"github.com/gin-gonic/gin"
)

// PaymentRoutes sets up all payment-related routes
func PaymentRoutes(r *gin.Engine) {
	pay := r.Group("/api/payment")
	{
		pay.GET("/providers", controllers.GetProviders)
		pay.POST("/providers", controllers.AddProvider)
		pay.POST("/verify", controllers.VerifyPayment)
	}
}

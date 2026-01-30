package main

import (
	"log"
	"payment_verification/routes"
	"payment_verification/utils"

	"github.com/gin-gonic/gin"
)
func main() {
	// Connect to MongoDB
	mongoURI := "mongodb://localhost:27017" 
	client, err := utils.ConnectDB(mongoURI)
	if err != nil {
		log.Fatal("Failed to connect to MongoDB:", err)
	}
	_ = client // optional, if you need to use it later

	// Create Gin router
	r := gin.Default()

	// Setup routes
	routes.PaymentRoutes(r)
	routes.UserRoutes(r)


	// Run server
	r.Run(":8081")
}

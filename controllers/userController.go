package controllers

import (
	"context"
	"net/http"
	"time"

	"payment_verification/models"
	"payment_verification/utils"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// AddUser creates a new user
func AddUser(c *gin.Context) {
	var user models.User
	if err := c.BindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user.ID = primitive.NewObjectID()
	user.CreatedAt = time.Now()
	

	collection := utils.GetCollection("payment_verification", "users")
	_, err := collection.InsertOne(context.Background(), user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add user"})
		return
	}

	c.JSON(http.StatusOK, user)
}

// GetUsers lists all users
func GetUsers(c *gin.Context) {
	collection := utils.GetCollection("payment_verification", "users")
	cursor, err := collection.Find(context.Background(), map[string]interface{}{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch users"})
		return
	}
	defer cursor.Close(context.Background())

	var users []models.User
	if err := cursor.All(context.Background(), &users); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse users"})
		return
	}

	c.JSON(http.StatusOK, users)
}

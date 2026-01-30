package controllers

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"payment_verification/models"
	"payment_verification/utils"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// AddProvider adds a new payment provider
func AddProvider(c *gin.Context) {
	var provider models.Provider
	if err := c.BindJSON(&provider); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	provider.ID = primitive.NewObjectID()
	collection := utils.GetCollection("payment_verification", "providers")
	_, err := collection.InsertOne(context.Background(), provider)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add provider"})
		return
	}

	c.JSON(http.StatusOK, provider)
}

// GetProviders lists all providers
func GetProviders(c *gin.Context) {
	collection := utils.GetCollection("payment_verification", "providers")
	cursor, err := collection.Find(context.Background(), bson.M{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch providers"})
		return
	}
	defer cursor.Close(context.Background())

	var providers []models.Provider
	if err := cursor.All(context.Background(), &providers); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse providers"})
		return
	}

	c.JSON(http.StatusOK, providers)
}

// VerifyPayment verifies a payment for a user
func VerifyPayment(c *gin.Context) {
	type Request struct {
		UserID     string `json:"user_id"`
		ProviderID string `json:"provider_id"`
		ReceiptID  string `json:"receipt_id"`
	}

	var req Request
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Convert ProviderID
	providerID, err := primitive.ObjectIDFromHex(req.ProviderID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid provider ID"})
		return
	}

	providerCollection := utils.GetCollection("payment_verification", "providers")
	var provider models.Provider
	err = providerCollection.FindOne(context.Background(), bson.M{"_id": providerID}).Decode(&provider)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Provider not found"})
		return
	}

	// Build full URL
	var fullURL string
	switch provider.Name {
	case "TeleBirr":
		fullURL = provider.URL + req.ReceiptID
	case "CBE":
		fullURL = provider.URL + "?id=" + req.ReceiptID
	default:
		fullURL = provider.URL + req.ReceiptID
	}

	fmt.Println("Fetching URL:", fullURL)

	// Fetch page with timeout
	client := http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(fullURL)
	if err != nil || resp.StatusCode != 200 {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch payment page"})
		return
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read page content"})
		return
	}
	bodyString := string(bodyBytes)

	// Extract payment details based on provider
	data := utils.ExtractPaymentData(bodyString, provider.Name)
	fmt.Println("Extracted Data:", data)
	fmt.Println("HTML Content (first 2000 chars):", bodyString[:min(len(bodyString), 2000)])

	// Convert UserID
	userID, err := primitive.ObjectIDFromHex(req.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// Save payment
payment := models.Payment{
    ID:                primitive.NewObjectID(),
    UserID:            userID,
    TotalAmount:       data["totalPaidAmount"],
    PaymentMode:       data["paymentMode"],
    PaymentReason:     data["paymentReason"],
    PayerName:         data["payerName"],
    PayerAccountNo:    data["payerAccountNo"],
    PaymentDate:       utils.ParsePaymentDate(data["paymentDate"]),
    InvoiceNo:         data["invoiceNo"],
    Status:            data["status"],
    ReceiverName:      data["creditedPartyName"],
    ReceiverAccountNo: data["creditedPartyAccountNo"],
    CreatedAt:         time.Now(),
}
	fmt.Println(payment)

	collection := utils.GetCollection("payment_verification", "payments")
	_, err = collection.InsertOne(context.Background(), payment)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save payment"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Payment verified and saved successfully",
		"payment": payment,
		"url":     fullURL,
	})
}

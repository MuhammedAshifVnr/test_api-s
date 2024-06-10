package handlers

import (
	"net/http"
	"test/db"
	"test/models"

	"github.com/gin-gonic/gin"
)

func AdminLogin(c *gin.Context) {
	type LoginInput struct {
		Email    string `json:"email" `
		Password string `json:"password" `
	}
	var input LoginInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	var Admin models.Admin
	if err := db.DB.Where("email = ?", input.Email).First(&Admin).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}
	if input.Password != Admin.Password {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Login successful"})
}

func GetAllUsers(c *gin.Context) {
    var users []models.User
    if err := db.DB.Find(&users).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch users"})
        return
    }
    c.JSON(http.StatusOK, gin.H{"users": users})
}
package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"github.com/syukurgit/zta/internal/domain"
	"github.com/syukurgit/zta/pkg/utils"
)

type AuthHandler struct {
	DB *gorm.DB
}

// Input struct untuk validasi JSON
type LoginInput struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

func (h *AuthHandler) Login(c *gin.Context) {
	var input LoginInput

	// 1. Validasi Input
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 2. Cari User di Database
	var user domain.User
	if err := h.DB.Where("email = ?", input.Email).First(&user).Error; err != nil {
		// PENTING: Jangan beri tahu jika email tidak ditemukan (prevent enumeration attack)
		// Katakan saja "Invalid email or password"
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}

	// 3. Cek Password
	if !utils.CheckPasswordHash(input.Password, user.PasswordHash) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}

	// 4. Generate Token (Berlaku 1 Jam)
	token, err := utils.GenerateToken(user.ID, user.Role, 1*time.Hour)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	// 5. Response
	c.JSON(http.StatusOK, gin.H{
		"token": token,
		"role":  user.Role, // Info tambahan untuk client
	})
}
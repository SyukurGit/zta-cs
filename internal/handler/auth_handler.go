package handler

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/syukurgit/zta/internal/domain"
	"github.com/syukurgit/zta/pkg/utils"
	"gorm.io/gorm"
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

	// 1. Bind JSON dulu
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 2. Cari user
	var user domain.User
	if err := h.DB.Where("email = ?", input.Email).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}

	// ===== DEBUG LOG (SEMENTARA) =====
	fmt.Println("====== LOGIN DEBUG ======")
	fmt.Println("EMAIL:", input.Email)
	fmt.Println("INPUT PASSWORD:", input.Password)
	fmt.Println("DB HASH:", user.PasswordHash)
	fmt.Println("PASSWORD MATCH:", utils.CheckPasswordHash(input.Password, user.PasswordHash))
	fmt.Println("=========================")
	// =================================

	// 3. Cek password
	if !utils.CheckPasswordHash(input.Password, user.PasswordHash) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}

	// 4. Generate token
	token, err := utils.GenerateToken(user.ID, user.Role, 1*time.Hour)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	// 5. Response
	c.JSON(http.StatusOK, gin.H{
		"token": token,
		"role":  user.Role,
	})
}

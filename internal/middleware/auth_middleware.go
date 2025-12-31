package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/syukurgit/zta/pkg/utils"
)

// AuthMiddleware memverifikasi Bearer Token
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. Ambil header Authorization
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			return
		}

		// 2. Format harus "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization format"})
			return
		}

		// 3. Validasi Token
		claims, err := utils.ValidateToken(parts[1])
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			return
		}

		// 4. Set Context (Identity Injection)
		// Simpan identitas ini agar bisa dipakai di Controller/Service nanti
		c.Set("user_id", claims.UserID)
		c.Set("role", claims.Role)

		c.Next() // Lanjut ke handler berikutnya
	}
}

// Tambahkan ini di internal/middleware/auth_middleware.go

func EnforceRole(allowedRole string) gin.HandlerFunc {
	return func(c *gin.Context) {
		role := c.GetString("role")
		if role != allowedRole {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Access denied: Insufficient privileges"})
			return
		}
		c.Next()
	}
}
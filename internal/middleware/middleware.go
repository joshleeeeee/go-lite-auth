package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/joshleeeeee/LiteSSO/internal/database"
	"github.com/joshleeeeee/LiteSSO/pkg/jwt"
)

// AuthMiddleware validates JWT token and sets user context
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := extractToken(c)
		if token == "" {
			c.JSON(401, gin.H{
				"code":    401,
				"message": "Authorization token required",
			})
			c.Abort()
			return
		}

		// Parse and validate token
		claims, err := jwt.ParseToken(token)
		if err != nil {
			c.JSON(401, gin.H{
				"code":    401,
				"message": "Invalid or expired token",
			})
			c.Abort()
			return
		}

		// Check if token type is access token
		if claims.Type != jwt.AccessToken {
			c.JSON(401, gin.H{
				"code":    401,
				"message": "Invalid token type",
			})
			c.Abort()
			return
		}

		// Check if token is blacklisted
		isBlacklisted, err := database.IsBlacklisted(c.Request.Context(), claims.TokenID)
		if err != nil || isBlacklisted {
			c.JSON(401, gin.H{
				"code":    401,
				"message": "Token has been revoked",
			})
			c.Abort()
			return
		}

		// Set user info in context
		c.Set("userID", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("claims", claims)

		c.Next()
	}
}

// CORSMiddleware handles CORS
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization")
		c.Header("Access-Control-Max-Age", "86400")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

// RecoveryMiddleware handles panics
func RecoveryMiddleware() gin.HandlerFunc {
	return gin.Recovery()
}

func extractToken(c *gin.Context) string {
	// First try Authorization header
	authHeader := c.GetHeader("Authorization")
	if authHeader != "" {
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) == 2 && parts[0] == "Bearer" {
			return parts[1]
		}
	}

	// Then try query parameter (for SSO redirect scenarios)
	if token := c.Query("token"); token != "" {
		return token
	}

	// Finally try cookie
	if token, err := c.Cookie("access_token"); err == nil {
		return token
	}

	return ""
}

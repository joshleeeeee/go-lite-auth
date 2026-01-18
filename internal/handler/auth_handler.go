package handler

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/joshleeeeee/LiteSSO/internal/service"
	"github.com/joshleeeeee/LiteSSO/pkg/jwt"
)

type AuthHandler struct {
	authService *service.AuthService
}

func NewAuthHandler() *AuthHandler {
	return &AuthHandler{
		authService: service.NewAuthService(),
	}
}

// Response represents a standard API response
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    0,
		Message: "success",
		Data:    data,
	})
}

func fail(c *gin.Context, code int, message string) {
	c.JSON(http.StatusOK, Response{
		Code:    code,
		Message: message,
	})
}

func unauthorized(c *gin.Context, message string) {
	c.JSON(http.StatusUnauthorized, Response{
		Code:    401,
		Message: message,
	})
}

// Register handles user registration
// POST /api/auth/register
func (h *AuthHandler) Register(c *gin.Context) {
	var req service.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, 400, "Invalid request: "+err.Error())
		return
	}

	user, err := h.authService.Register(c.Request.Context(), &req)
	if err != nil {
		fail(c, 400, err.Error())
		return
	}

	success(c, gin.H{
		"user_id":  user.ID,
		"username": user.Username,
		"email":    user.Email,
	})
}

// Login handles user login
// POST /api/auth/login
func (h *AuthHandler) Login(c *gin.Context) {
	var req service.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, 400, "Invalid request: "+err.Error())
		return
	}

	clientIP := c.ClientIP()
	resp, err := h.authService.Login(c.Request.Context(), &req, clientIP)
	if err != nil {
		fail(c, 401, err.Error())
		return
	}

	success(c, resp)
}

// Logout handles user logout
// POST /api/auth/logout
func (h *AuthHandler) Logout(c *gin.Context) {
	token := extractToken(c)
	if token == "" {
		unauthorized(c, "No token provided")
		return
	}

	if err := h.authService.Logout(c.Request.Context(), token); err != nil {
		fail(c, 500, "Logout failed")
		return
	}

	success(c, nil)
}

// RefreshToken refreshes the access token
// POST /api/auth/refresh
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, 400, "Invalid request")
		return
	}

	tokenPair, err := h.authService.RefreshToken(c.Request.Context(), req.RefreshToken)
	if err != nil {
		unauthorized(c, "Invalid refresh token")
		return
	}

	success(c, tokenPair)
}

// GetUserInfo returns the current user's information
// GET /api/user/info
func (h *AuthHandler) GetUserInfo(c *gin.Context) {
	// Get user ID from context (set by auth middleware)
	userID, exists := c.Get("userID")
	if !exists {
		unauthorized(c, "Unauthorized")
		return
	}

	user, err := h.authService.GetUserInfo(c.Request.Context(), userID.(uint))
	if err != nil {
		fail(c, 404, "User not found")
		return
	}

	success(c, user)
}

// ValidateToken validates the current token
// GET /api/auth/validate
func (h *AuthHandler) ValidateToken(c *gin.Context) {
	token := extractToken(c)
	if token == "" {
		unauthorized(c, "No token provided")
		return
	}

	claims, err := h.authService.ValidateToken(c.Request.Context(), token)
	if err != nil {
		unauthorized(c, "Invalid token")
		return
	}

	success(c, gin.H{
		"user_id":  claims.UserID,
		"username": claims.Username,
		"valid":    true,
	})
}

// extractToken extracts the JWT token from Authorization header
func extractToken(c *gin.Context) string {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		return ""
	}

	// Expected format: Bearer <token>
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || parts[0] != "Bearer" {
		return ""
	}

	return parts[1]
}

// GetClaims helper to get claims from context
func GetClaims(c *gin.Context) *jwt.Claims {
	claims, exists := c.Get("claims")
	if !exists {
		return nil
	}
	return claims.(*jwt.Claims)
}

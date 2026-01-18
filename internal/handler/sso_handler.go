package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/joshleeeeee/go-lite-auth/internal/service"
)

// SSOHandler handles SSO-related HTTP requests
type SSOHandler struct {
	ssoService *service.SSOService
}

// NewSSOHandler creates a new SSOHandler instance
func NewSSOHandler() *SSOHandler {
	return &SSOHandler{
		ssoService: service.NewSSOService(),
	}
}

// Login handles SSO login requests
// GET /sso/login?service=xxx - Check login state, redirect if already logged in
// POST /sso/login - Process login form, generate ST and redirect
func (h *SSOHandler) Login(c *gin.Context) {
	serviceURL := c.Query("service")
	if serviceURL == "" {
		c.JSON(http.StatusBadRequest, Response{
			Code:    400,
			Message: "Missing required parameter: service",
		})
		return
	}

	// For GET request, show login page info (or redirect if already logged in)
	// In a real implementation, you would check for TGT cookie here
	c.JSON(http.StatusOK, Response{
		Code:    0,
		Message: "Please login",
		Data: gin.H{
			"service":   serviceURL,
			"login_url": "/sso/login",
		},
	})
}

// LoginSubmit handles SSO login form submission
// POST /sso/login
func (h *SSOHandler) LoginSubmit(c *gin.Context) {
	var req service.SSOLoginRequest

	// Support both JSON and form data
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Code:    400,
			Message: "Invalid request: " + err.Error(),
		})
		return
	}

	clientIP := c.ClientIP()
	resp, err := h.ssoService.Login(c.Request.Context(), &req, clientIP)
	if err != nil {
		statusCode := http.StatusUnauthorized
		if err == service.ErrInvalidService {
			statusCode = http.StatusBadRequest
		}
		c.JSON(statusCode, Response{
			Code:    statusCode,
			Message: err.Error(),
		})
		return
	}

	// Return ticket and redirect URL
	// Client applications should redirect to resp.RedirectURL
	c.JSON(http.StatusOK, Response{
		Code:    0,
		Message: "Login successful",
		Data:    resp,
	})
}

// ValidateTicket validates a Service Ticket and returns user info
// GET /sso/validate?ticket=xxx&service=xxx
func (h *SSOHandler) ValidateTicket(c *gin.Context) {
	ticket := c.Query("ticket")
	serviceURL := c.Query("service")

	if ticket == "" {
		c.JSON(http.StatusBadRequest, Response{
			Code:    400,
			Message: "Missing required parameter: ticket",
		})
		return
	}

	if serviceURL == "" {
		c.JSON(http.StatusBadRequest, Response{
			Code:    400,
			Message: "Missing required parameter: service",
		})
		return
	}

	userInfo, err := h.ssoService.ValidateServiceTicket(c.Request.Context(), ticket, serviceURL)
	if err != nil {
		statusCode := http.StatusUnauthorized
		message := "Invalid or expired ticket"

		switch err {
		case service.ErrTicketNotFound:
			message = "Ticket not found or expired"
		case service.ErrServiceMismatch:
			message = "Service URL mismatch"
			statusCode = http.StatusBadRequest
		}

		c.JSON(statusCode, Response{
			Code:    statusCode,
			Message: message,
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Code:    0,
		Message: "Ticket validated successfully",
		Data:    userInfo,
	})
}

// Logout handles SSO logout
// GET /sso/logout?service=xxx
func (h *SSOHandler) Logout(c *gin.Context) {
	serviceURL := c.Query("service")

	// In a full implementation, you would:
	// 1. Clear TGT cookie
	// 2. Notify all registered services to clear their sessions (SLO)
	// 3. Redirect to service URL if provided

	c.JSON(http.StatusOK, Response{
		Code:    0,
		Message: "Logged out successfully",
		Data: gin.H{
			"redirect_url": serviceURL,
		},
	})
}

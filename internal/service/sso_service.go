package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/joshleeeeee/go-lite-auth/internal/database"
	"github.com/joshleeeeee/go-lite-auth/internal/model"
	"github.com/joshleeeeee/go-lite-auth/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

// SSO-related errors
var (
	ErrTicketNotFound   = errors.New("ticket not found or expired")
	ErrTicketUsed       = errors.New("ticket has already been used")
	ErrServiceMismatch  = errors.New("service URL mismatch")
	ErrInvalidService   = errors.New("invalid or missing service URL")
)

// Ticket configuration
const (
	TicketPrefix   = "ST-"
	TicketExpire   = 60 * time.Second // Service Ticket expires in 60 seconds
	TicketIDLength = 32               // Length of random ticket ID
)

// SSOService handles SSO-related business logic
type SSOService struct {
	userRepo    *repository.UserRepository
	authService *AuthService
}

// NewSSOService creates a new SSOService instance
func NewSSOService() *SSOService {
	return &SSOService{
		userRepo:    repository.NewUserRepository(),
		authService: NewAuthService(),
	}
}

// SSOLoginRequest represents an SSO login request
type SSOLoginRequest struct {
	Username string `json:"username" form:"username" binding:"required"`
	Password string `json:"password" form:"password" binding:"required"`
	Service  string `json:"service" form:"service" binding:"required"`
}

// SSOLoginResponse represents the response after successful SSO login
type SSOLoginResponse struct {
	Ticket      string `json:"ticket"`
	RedirectURL string `json:"redirect_url"`
}

// ValidateTicketResponse represents the response when validating a ticket
type ValidateTicketResponse struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Nickname string `json:"nickname"`
}

// generateTicketID generates a random ticket ID
func generateTicketID() (string, error) {
	bytes := make([]byte, TicketIDLength/2)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return TicketPrefix + hex.EncodeToString(bytes), nil
}

// Login authenticates a user and generates a Service Ticket for SSO
func (s *SSOService) Login(ctx context.Context, req *SSOLoginRequest, clientIP string) (*SSOLoginResponse, error) {
	if req.Service == "" {
		return nil, ErrInvalidService
	}

	// Check login rate limit
	failKey := fmt.Sprintf("%s:%s", clientIP, req.Username)
	failCount, err := database.GetLoginFailCount(ctx, failKey)
	if err != nil {
		return nil, err
	}
	if failCount >= MaxLoginAttempts {
		return nil, ErrTooManyAttempts
	}

	// Find user
	user, err := s.userRepo.GetByUsername(req.Username)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			database.IncrLoginFail(ctx, failKey, LoginLockDuration)
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}

	// Check password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		database.IncrLoginFail(ctx, failKey, LoginLockDuration)
		return nil, ErrInvalidCredentials
	}

	// Check user status
	if user.Status != 1 {
		return nil, ErrUserDisabled
	}

	// Clear login failures on success
	database.ClearLoginFail(ctx, failKey)

	// Generate Service Ticket
	ticket, err := s.GenerateServiceTicket(ctx, user, req.Service)
	if err != nil {
		return nil, fmt.Errorf("failed to generate ticket: %w", err)
	}

	// Build redirect URL with ticket
	redirectURL := buildRedirectURL(req.Service, ticket)

	return &SSOLoginResponse{
		Ticket:      ticket,
		RedirectURL: redirectURL,
	}, nil
}

// GenerateServiceTicket creates a one-time Service Ticket
func (s *SSOService) GenerateServiceTicket(ctx context.Context, user *model.User, service string) (string, error) {
	ticketID, err := generateTicketID()
	if err != nil {
		return "", fmt.Errorf("failed to generate ticket ID: %w", err)
	}

	ticketData := &database.TicketData{
		UserID:   user.ID,
		Username: user.Username,
		Service:  service,
	}

	if err := database.SetTicketWithService(ctx, ticketID, ticketData, TicketExpire); err != nil {
		return "", fmt.Errorf("failed to store ticket: %w", err)
	}

	return ticketID, nil
}

// ValidateServiceTicket validates and consumes a Service Ticket (one-time use)
func (s *SSOService) ValidateServiceTicket(ctx context.Context, ticket, service string) (*ValidateTicketResponse, error) {
	if ticket == "" {
		return nil, ErrTicketNotFound
	}
	if service == "" {
		return nil, ErrInvalidService
	}

	// Atomically get and delete ticket (ensures one-time use)
	ticketData, err := database.GetAndDeleteTicketData(ctx, ticket)
	if err != nil {
		return nil, ErrTicketNotFound
	}

	// Validate service URL matches
	if ticketData.Service != service {
		return nil, ErrServiceMismatch
	}

	// Get full user info
	user, err := s.userRepo.GetByID(ticketData.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}

	return &ValidateTicketResponse{
		UserID:   user.ID,
		Username: user.Username,
		Email:    user.Email,
		Nickname: user.Nickname,
	}, nil
}

// buildRedirectURL appends the ticket parameter to the service URL
func buildRedirectURL(service, ticket string) string {
	separator := "?"
	// Check if service URL already has query parameters
	for _, c := range service {
		if c == '?' {
			separator = "&"
			break
		}
	}
	return service + separator + "ticket=" + ticket
}

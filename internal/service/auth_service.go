package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/joshleeeeee/LiteSSO/internal/config"
	"github.com/joshleeeeee/LiteSSO/internal/database"
	"github.com/joshleeeeee/LiteSSO/internal/model"
	"github.com/joshleeeeee/LiteSSO/internal/repository"
	"github.com/joshleeeeee/LiteSSO/pkg/jwt"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid username or password")
	ErrUserDisabled       = errors.New("user account is disabled")
	ErrTooManyAttempts    = errors.New("too many login attempts, please try again later")
)

const (
	MaxLoginAttempts  = 5
	LoginLockDuration = 5 * time.Minute
)

type AuthService struct {
	userRepo *repository.UserRepository
}

func NewAuthService() *AuthService {
	return &AuthService{
		userRepo: repository.NewUserRepository(),
	}
}

// RegisterRequest represents a registration request
type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6,max=50"`
	Nickname string `json:"nickname"`
}

// LoginRequest represents a login request
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// AuthResponse represents the authentication response
type AuthResponse struct {
	User      *model.User    `json:"user"`
	TokenPair *jwt.TokenPair `json:"token"`
}

// Register creates a new user account
func (s *AuthService) Register(ctx context.Context, req *RegisterRequest) (*model.User, error) {
	// Check if username exists
	exists, err := s.userRepo.ExistsByUsername(req.Username)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, fmt.Errorf("username '%s' already exists", req.Username)
	}

	// Check if email exists
	exists, err = s.userRepo.ExistsByEmail(req.Email)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, fmt.Errorf("email '%s' already exists", req.Email)
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create user
	user := &model.User{
		Username: req.Username,
		Email:    req.Email,
		Password: string(hashedPassword),
		Nickname: req.Nickname,
		Status:   1,
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

// Login authenticates a user and returns tokens
func (s *AuthService) Login(ctx context.Context, req *LoginRequest, clientIP string) (*AuthResponse, error) {
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

	// Generate tokens
	tokenPair, err := jwt.GenerateTokenPair(user.ID, user.Username)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	// Store session in Redis
	sessionExpire := config.GlobalConfig.Session.ExpireDuration()
	if err := database.SetSession(ctx, tokenPair.AccessToken[:32], user.ID, sessionExpire); err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	return &AuthResponse{
		User:      user,
		TokenPair: tokenPair,
	}, nil
}

// Logout invalidates the current session and token
func (s *AuthService) Logout(ctx context.Context, tokenString string) error {
	// Parse token to get claims
	claims, err := jwt.ParseToken(tokenString)
	if err != nil {
		return err
	}

	// Add token to blacklist
	remainingTime := jwt.GetTokenRemainingTime(claims)
	if remainingTime > 0 {
		if err := database.AddToBlacklist(ctx, claims.TokenID, remainingTime); err != nil {
			return fmt.Errorf("failed to blacklist token: %w", err)
		}
	}

	// Delete session
	database.DeleteSession(ctx, tokenString[:32])

	return nil
}

// RefreshToken generates new token pair using refresh token
func (s *AuthService) RefreshToken(ctx context.Context, refreshToken string) (*jwt.TokenPair, error) {
	// Parse refresh token
	claims, err := jwt.ParseToken(refreshToken)
	if err != nil {
		return nil, err
	}

	// Verify it's a refresh token
	if claims.Type != jwt.RefreshToken {
		return nil, jwt.ErrInvalidToken
	}

	// Check if token is blacklisted
	isBlacklisted, err := database.IsBlacklisted(ctx, claims.TokenID)
	if err != nil {
		return nil, err
	}
	if isBlacklisted {
		return nil, jwt.ErrInvalidToken
	}

	// Blacklist old refresh token
	remainingTime := jwt.GetTokenRemainingTime(claims)
	if remainingTime > 0 {
		database.AddToBlacklist(ctx, claims.TokenID, remainingTime)
	}

	// Generate new token pair
	return jwt.GenerateTokenPair(claims.UserID, claims.Username)
}

// ValidateToken validates an access token
func (s *AuthService) ValidateToken(ctx context.Context, tokenString string) (*jwt.Claims, error) {
	claims, err := jwt.ParseToken(tokenString)
	if err != nil {
		return nil, err
	}

	// Check if token is blacklisted
	isBlacklisted, err := database.IsBlacklisted(ctx, claims.TokenID)
	if err != nil {
		return nil, err
	}
	if isBlacklisted {
		return nil, jwt.ErrInvalidToken
	}

	return claims, nil
}

// GetUserInfo returns user information
func (s *AuthService) GetUserInfo(ctx context.Context, userID uint) (*model.User, error) {
	return s.userRepo.GetByID(userID)
}

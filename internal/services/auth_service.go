package services

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/tldr-it-stepankutaj/openvpn-mng/internal/config"
	"github.com/tldr-it-stepankutaj/openvpn-mng/internal/database"
	"github.com/tldr-it-stepankutaj/openvpn-mng/internal/middleware"
	"github.com/tldr-it-stepankutaj/openvpn-mng/internal/models"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid username or password")
	ErrUserNotFound       = errors.New("user not found")
	ErrUserInactive       = errors.New("user account is inactive")
	ErrUserNotYetValid    = errors.New("user account is not yet valid")
	ErrUserExpired        = errors.New("user account has expired")
)

// AuthService provides authentication services
type AuthService struct {
	config *config.AuthConfig
}

// NewAuthService creates a new auth service
func NewAuthService(cfg *config.AuthConfig) *AuthService {
	return &AuthService{
		config: cfg,
	}
}

// Authenticate authenticates a user and returns a JWT token
func (s *AuthService) Authenticate(username, password string) (string, *models.User, error) {
	var user models.User
	if err := database.GetDB().Where("username = ?", username).First(&user).Error; err != nil {
		return "", nil, ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return "", nil, ErrInvalidCredentials
	}

	// Check if user is active and within validity period
	if err := s.validateUserAccess(&user); err != nil {
		return "", nil, err
	}

	// Generate JWT token
	token, err := s.generateToken(&user)
	if err != nil {
		return "", nil, err
	}

	return token, &user, nil
}

// validateUserAccess checks if user is active and within validity period
func (s *AuthService) validateUserAccess(user *models.User) error {
	if !user.IsActive {
		return ErrUserInactive
	}

	if !user.IsValidForLogin() {
		// Determine specific error
		now := time.Now()
		today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

		if user.ValidFrom != nil {
			validFromDate := time.Date(user.ValidFrom.Year(), user.ValidFrom.Month(), user.ValidFrom.Day(), 0, 0, 0, 0, user.ValidFrom.Location())
			if today.Before(validFromDate) {
				return ErrUserNotYetValid
			}
		}

		if user.ValidTo != nil {
			validToDate := time.Date(user.ValidTo.Year(), user.ValidTo.Month(), user.ValidTo.Day(), 23, 59, 59, 0, user.ValidTo.Location())
			if today.After(validToDate) {
				return ErrUserExpired
			}
		}
	}

	return nil
}

// generateToken generates a JWT token for a user
func (s *AuthService) generateToken(user *models.User) (string, error) {
	claims := &middleware.Claims{
		UserID:   user.ID.String(),
		Username: user.Username,
		Role:     user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(s.config.TokenExpiry) * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.config.JWTSecret))
}

// HashPassword hashes a password
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// VerifyPassword verifies a password against a hash
func VerifyPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// GetUserByID returns a user by ID
func GetUserByID(id uuid.UUID) (*models.User, error) {
	var user models.User
	if err := database.GetDB().Preload("Manager").First(&user, "id = ?", id).Error; err != nil {
		return nil, ErrUserNotFound
	}
	return &user, nil
}

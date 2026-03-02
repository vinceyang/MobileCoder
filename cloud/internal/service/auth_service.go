package service

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"

	"github.com/mobile-coder/cloud/internal/db"
)

var ErrInvalidCredentials = errors.New("invalid email or password")
var ErrUserAlreadyExists = errors.New("user already exists")

type AuthService struct {
	db *db.SupabaseDB
}

func NewAuthService(database *db.SupabaseDB) *AuthService {
	return &AuthService{db: database}
}

func hashPassword(password string) string {
	hash := sha256.Sum256([]byte(password))
	return hex.EncodeToString(hash[:])
}

// Register creates a new user
func (s *AuthService) Register(email, password string) (*db.User, error) {
	// Check if user already exists
	existingUser, _ := s.db.GetUserByEmail(email)
	if existingUser != nil {
		return nil, ErrUserAlreadyExists
	}

	hashedPassword := hashPassword(password)
	// Use email as username for simplicity
	return s.db.CreateUser(email, hashedPassword, email)
}

// Login validates credentials and returns user
func (s *AuthService) Login(email, password string) (*db.User, error) {
	user, err := s.db.GetUserByEmail(email)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	hashedPassword := hashPassword(password)
	if user.Password != hashedPassword {
		return nil, ErrInvalidCredentials
	}

	return user, nil
}

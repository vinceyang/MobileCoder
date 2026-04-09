package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrExpiredToken = errors.New("expired token")
)

type Claims struct {
	UserID    int64  `json:"user_id"`
	Email     string `json:"email"`
	ExpiresAt int64  `json:"expires_at"`
}

type Manager struct {
	secret []byte
	ttl    time.Duration
	now    func() time.Time
}

func NewManager(secret string, ttl time.Duration) *Manager {
	return &Manager{
		secret: []byte(secret),
		ttl:    ttl,
		now:    time.Now,
	}
}

func (m *Manager) Issue(userID int64, email string) (string, error) {
	claims := Claims{
		UserID:    userID,
		Email:     email,
		ExpiresAt: m.now().Add(m.ttl).Unix(),
	}

	payload, err := json.Marshal(claims)
	if err != nil {
		return "", err
	}

	encodedPayload := base64.RawURLEncoding.EncodeToString(payload)
	signature := m.sign(encodedPayload)
	return encodedPayload + "." + signature, nil
}

func (m *Manager) Verify(token string) (*Claims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 2 {
		return nil, ErrInvalidToken
	}

	if !hmac.Equal([]byte(parts[1]), []byte(m.sign(parts[0]))) {
		return nil, ErrInvalidToken
	}

	payload, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return nil, ErrInvalidToken
	}

	var claims Claims
	if err := json.Unmarshal(payload, &claims); err != nil {
		return nil, ErrInvalidToken
	}

	if claims.UserID == 0 || claims.Email == "" {
		return nil, ErrInvalidToken
	}

	if m.now().Unix() > claims.ExpiresAt {
		return nil, ErrExpiredToken
	}

	return &claims, nil
}

func (m *Manager) sign(payload string) string {
	mac := hmac.New(sha256.New, m.secret)
	_, _ = mac.Write([]byte(payload))
	return fmt.Sprintf("%s", hex.EncodeToString(mac.Sum(nil)))
}

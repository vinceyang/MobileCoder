package handler

import (
	"errors"
	"net/http"

	cloudauth "github.com/mobile-coder/cloud/internal/auth"
	"github.com/mobile-coder/cloud/internal/service"
)

var (
	errUnauthorized = errors.New("unauthorized")
	errForbidden    = errors.New("forbidden")
)

func requireClaims(token string, manager *cloudauth.Manager) (*cloudauth.Claims, error) {
	if token == "" || manager == nil {
		return nil, errUnauthorized
	}

	claims, err := manager.Verify(token)
	if err != nil {
		return nil, errUnauthorized
	}
	return claims, nil
}

func requireClaimsFromRequest(r *http.Request, manager *cloudauth.Manager) (*cloudauth.Claims, error) {
	return requireClaims(r.Header.Get("Authorization"), manager)
}

func ensureDeviceOwnership(device *service.Device, userID int64) error {
	if device == nil || device.UserID == 0 || device.UserID != userID {
		return errForbidden
	}
	return nil
}

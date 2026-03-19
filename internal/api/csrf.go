package api

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"strings"
)

const (
	authCookieName = "openfiltr_token"
	csrfCookieName = "openfiltr_csrf"
	csrfHeaderName = "X-CSRF-Token"
)

func newCSRFToken() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}

func isStateChanging(method string) bool {
	switch method {
	case http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete:
		return true
	default:
		return false
	}
}

func bearerTokenFromRequest(r *http.Request) string {
	authz := r.Header.Get("Authorization")
	if authz == "" {
		return ""
	}
	parts := strings.SplitN(authz, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
		return ""
	}
	return parts[1]
}

func setCSRFCookie(w http.ResponseWriter, value string, maxAge int) {
	http.SetCookie(w, &http.Cookie{
		Name:     csrfCookieName,
		Value:    value,
		Path:     "/",
		MaxAge:   maxAge,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	})
}

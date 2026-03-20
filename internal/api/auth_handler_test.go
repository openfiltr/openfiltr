package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/openfiltr/openfiltr/internal/auth"
	"github.com/openfiltr/openfiltr/internal/config"
)

func TestLoginSetsCSRFHeaderAndCookie(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New() error = %v", err)
	}
	defer db.Close()

	hash, err := auth.HashPassword("correct horse battery staple")
	if err != nil {
		t.Fatalf("HashPassword() error = %v", err)
	}

	rows := sqlmock.NewRows([]string{"id", "password_hash", "role"}).AddRow("user-1", hash, "admin")
	mock.ExpectQuery(`SELECT id,password_hash,role FROM users WHERE username=\$1`).
		WithArgs("alice").
		WillReturnRows(rows)

	h := &Handler{
		db:      db,
		cfg:     &config.Config{Auth: config.Auth{JWTSecret: "test-secret", TokenExpiry: 2}},
		authSvc: auth.NewService(db, "test-secret", 2),
	}

	body, _ := json.Marshal(map[string]string{"username": "alice", "password": "correct horse battery staple"})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(body))
	w := httptest.NewRecorder()

	h.Login(w, req)

	res := w.Result()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("Login() status = %d, want %d", res.StatusCode, http.StatusOK)
	}
	csrfHeader := res.Header.Get(csrfHeaderName)
	if csrfHeader == "" {
		t.Fatal("Login() missing X-CSRF-Token header")
	}

	var sessionCookie, csrfCookie *http.Cookie
	for _, cookie := range res.Cookies() {
		switch cookie.Name {
		case authCookieName:
			sessionCookie = cookie
		case csrfCookieName:
			csrfCookie = cookie
		}
	}
	if sessionCookie == nil {
		t.Fatal("Login() missing session cookie")
	}
	if csrfCookie == nil {
		t.Fatal("Login() missing CSRF cookie")
	}
	if csrfCookie.Value != csrfHeader {
		t.Fatalf("CSRF cookie = %q, want %q", csrfCookie.Value, csrfHeader)
	}
	if csrfCookie.HttpOnly {
		t.Fatal("CSRF cookie should remain readable by the browser")
	}
	if csrfCookie.SameSite != http.SameSiteStrictMode {
		t.Fatalf("CSRF cookie SameSite = %v, want %v", csrfCookie.SameSite, http.SameSiteStrictMode)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations not met: %v", err)
	}
}

func TestBrowserSessionStateChangeRequiresCSRFToken(t *testing.T) {
	h := newCSRFMiddlewareTestHandler(t, nil)
	cookieToken := mustSignedToken(t, h.authSvc)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/tokens", nil)
	req.AddCookie(&http.Cookie{Name: authCookieName, Value: cookieToken})
	req.AddCookie(&http.Cookie{Name: csrfCookieName, Value: "csrf-token"})
	w := httptest.NewRecorder()

	h.AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})).ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusForbidden)
	}
}

func TestBrowserSessionStateChangeAcceptsMatchingCSRFToken(t *testing.T) {
	h := newCSRFMiddlewareTestHandler(t, nil)
	cookieToken := mustSignedToken(t, h.authSvc)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/tokens", nil)
	req.AddCookie(&http.Cookie{Name: authCookieName, Value: cookieToken})
	req.AddCookie(&http.Cookie{Name: csrfCookieName, Value: "csrf-token"})
	req.Header.Set(csrfHeaderName, "csrf-token")
	w := httptest.NewRecorder()

	h.AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})).ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusNoContent)
	}
}

func TestBearerAuthBypassesCSRFCheck(t *testing.T) {
	h := newCSRFMiddlewareTestHandler(t, nil)
	bearerToken := mustSignedToken(t, h.authSvc)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/tokens", nil)
	req.Header.Set("Authorization", "Bearer "+bearerToken)
	w := httptest.NewRecorder()

	h.AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})).ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusNoContent)
	}
}

func TestLogoutClearsCSRFCookie(t *testing.T) {
	h := &Handler{}
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout", nil)
	w := httptest.NewRecorder()

	h.Logout(w, req)

	res := w.Result()
	var csrfCookie *http.Cookie
	for _, cookie := range res.Cookies() {
		if cookie.Name == csrfCookieName {
			csrfCookie = cookie
			break
		}
	}
	if csrfCookie == nil {
		t.Fatal("Logout() missing CSRF cookie reset")
	}
	if csrfCookie.MaxAge != -1 {
		t.Fatalf("CSRF cookie MaxAge = %d, want -1", csrfCookie.MaxAge)
	}
}

func TestNewRouterAllowsAndExposesCSRFHeader(t *testing.T) {
	r := NewRouter(&config.Config{Auth: config.Auth{JWTSecret: "secret", TokenExpiry: 1}}, &sql.DB{}, "test")

	preflightReq := httptest.NewRequest(http.MethodOptions, "/api/v1/auth/login", nil)
	preflightReq.Header.Set("Origin", "https://example.com")
	preflightReq.Header.Set("Access-Control-Request-Method", http.MethodPost)
	preflightReq.Header.Set("Access-Control-Request-Headers", csrfHeaderName)
	preflight := httptest.NewRecorder()

	r.ServeHTTP(preflight, preflightReq)

	allowHeaders := preflight.Header().Get("Access-Control-Allow-Headers")
	if allowHeaders == "" || !containsToken(allowHeaders, csrfHeaderName) {
		t.Fatalf("Access-Control-Allow-Headers = %q, want %q present", allowHeaders, csrfHeaderName)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/system/health", nil)
	req.Header.Set("Origin", "https://example.com")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	exposeHeaders := w.Header().Get("Access-Control-Expose-Headers")
	if exposeHeaders == "" || !containsToken(exposeHeaders, csrfHeaderName) {
		t.Fatalf("Access-Control-Expose-Headers = %q, want %q present", exposeHeaders, csrfHeaderName)
	}
}

func newCSRFMiddlewareTestHandler(t *testing.T, db *sql.DB) *Handler {
	t.Helper()
	if db == nil {
		db = &sql.DB{}
	}
	return &Handler{
		db:      db,
		cfg:     &config.Config{Auth: config.Auth{JWTSecret: "test-secret", TokenExpiry: 1}},
		authSvc: auth.NewService(db, "test-secret", 1),
	}
}

func mustSignedToken(t *testing.T, svc *auth.Service) string {
	t.Helper()
	token, err := svc.GenerateToken("user-1", "alice", "admin")
	if err != nil {
		t.Fatalf("GenerateToken() error = %v", err)
	}
	return token
}

func containsToken(header, want string) bool {
	for _, part := range strings.Split(header, ",") {
		part = strings.TrimSpace(part)
		if strings.EqualFold(part, want) {
			return true
		}
	}
	return false
}

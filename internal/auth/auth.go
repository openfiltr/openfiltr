package auth

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type Claims struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

type Service struct {
	db        *sql.DB
	jwtSecret []byte
	expiry    time.Duration
}

func NewService(db *sql.DB, secret string, hours int) *Service {
	return &Service{db: db, jwtSecret: []byte(secret), expiry: time.Duration(hours) * time.Hour}
}

func HashPassword(pw string) (string, error) {
	h, err := bcrypt.GenerateFromPassword([]byte(pw), bcrypt.DefaultCost)
	return string(h), err
}

func CheckPassword(pw, hash string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(pw)) == nil
}

func (s *Service) GenerateToken(userID, username, role string) (string, error) {
	claims := Claims{
		UserID: userID, Username: username, Role: role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.expiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   userID,
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(s.jwtSecret)
}

func (s *Service) ValidateToken(tokenStr string) (*Claims, error) {
	t, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return s.jwtSecret, nil
	})
	if err != nil {
		return nil, err
	}
	c, ok := t.Claims.(*Claims)
	if !ok || !t.Valid {
		return nil, fmt.Errorf("invalid token")
	}
	return c, nil
}

func GenerateAPIToken() (raw, hash string, err error) {
	b := make([]byte, 32)
	if _, err = rand.Read(b); err != nil {
		return
	}
	raw = "oft_" + hex.EncodeToString(b)
	h, err := bcrypt.GenerateFromPassword([]byte(raw), bcrypt.MinCost)
	hash = string(h)
	return
}

func (s *Service) ValidateAPIToken(token string) (*Claims, error) {
	rows, err := s.db.Query(`SELECT at.token_hash,u.id,u.username,u.role FROM api_tokens at JOIN users u ON u.id=at.user_id WHERE (at.expires_at IS NULL OR at.expires_at>datetime('now'))`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var h, uid, uname, role string
		if err := rows.Scan(&h, &uid, &uname, &role); err != nil {
			continue
		}
		if bcrypt.CompareHashAndPassword([]byte(h), []byte(token)) == nil {
			_, _ = s.db.Exec("UPDATE api_tokens SET last_used_at=datetime('now') WHERE token_hash=?", h)
			return &Claims{UserID: uid, Username: uname, Role: role}, nil
		}
	}
	return nil, fmt.Errorf("invalid API token")
}

func EnsureAdminUser(db *sql.DB, username, password string) error {
	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM users").Scan(&count); err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	hash, err := HashPassword(password)
	if err != nil {
		return err
	}
	_, err = db.Exec(`INSERT INTO users(id,username,email,password_hash,role) VALUES(?,?,?,?,'admin')`,
		uuid.New().String(), username, username+"@localhost", hash)
	return err
}

func ExtractToken(r *http.Request) string {
	if h := r.Header.Get("Authorization"); h != "" {
		if parts := strings.SplitN(h, " ", 2); len(parts) == 2 && strings.EqualFold(parts[0], "bearer") {
			return parts[1]
		}
	}
	if c, err := r.Cookie("openfiltr_token"); err == nil {
		return c.Value
	}
	return ""
}

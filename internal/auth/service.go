package auth

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/openfiltr/openfiltr/internal/storage"
)

type UserRecord struct {
	ID           string
	Username     string
	PasswordHash string
	Role         string
}

type APITokenView struct {
	ID         string  `json:"id"`
	Name       string  `json:"name"`
	Scopes     string  `json:"scopes"`
	LastUsedAt *string `json:"last_used_at"`
	ExpiresAt  *string `json:"expires_at"`
	CreatedAt  string  `json:"created_at"`
}

type authRepository interface {
	LookupUserByUsername(username string) (UserRecord, error)
	LookupUserByID(id string) (UserRecord, error)
	CountUsers() (int, error)
	CreateAdminUser(id, username, email, passwordHash string) error
	ListAPITokens(userID string) ([]APITokenView, error)
	CreateAPIToken(id, userID, name, tokenHash string, expiresAt *time.Time) error
	DeleteAPIToken(id, userID string) (bool, error)
	ValidateAPIToken(token string) (*Claims, error)
}

type Service struct {
	repo      authRepository
	jwtSecret []byte
	expiry    time.Duration
}

func NewService(db storage.Store, secret string, hours int) *Service {
	return &Service{
		repo:      newAuthRepository(db),
		jwtSecret: []byte(secret),
		expiry:    time.Duration(hours) * time.Hour,
	}
}

func (s *Service) LookupUserByUsername(username string) (UserRecord, error) {
	return s.repo.LookupUserByUsername(username)
}

func (s *Service) LookupUserByID(id string) (UserRecord, error) {
	return s.repo.LookupUserByID(id)
}

func (s *Service) CountUsers() (int, error) {
	return s.repo.CountUsers()
}

func (s *Service) CreateAdminUser(username, password string) error {
	hash, err := HashPassword(password)
	if err != nil {
		return err
	}
	return s.repo.CreateAdminUser(uuid.New().String(), username, username+"@localhost", hash)
}

func (s *Service) ListAPITokens(userID string) ([]APITokenView, error) {
	items, err := s.repo.ListAPITokens(userID)
	if err != nil {
		return nil, err
	}
	if items == nil {
		return []APITokenView{}, nil
	}
	return items, nil
}

func (s *Service) CreateAPIToken(userID, name, tokenHash string, expiresAt *time.Time) (string, error) {
	id := uuid.New().String()
	if err := s.repo.CreateAPIToken(id, userID, name, tokenHash, expiresAt); err != nil {
		return "", err
	}
	return id, nil
}

func (s *Service) DeleteAPIToken(id, userID string) (bool, error) {
	return s.repo.DeleteAPIToken(id, userID)
}

func (s *Service) ValidateAPIToken(token string) (*Claims, error) {
	return s.repo.ValidateAPIToken(token)
}

func (s *Service) EnsureAdminUser(username, password string) error {
	count, err := s.CountUsers()
	if err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	return s.CreateAdminUser(username, password)
}

func newAuthRepository(db storage.Store) authRepository {
	if bolt, ok := db.(*storage.BoltStore); ok {
		return newBoltRepository(bolt)
	}
	return newSQLRepository(db)
}

func formatTimeText(t time.Time) string {
	return t.UTC().Format(time.RFC3339)
}

func formatOptionalTimeText(t *time.Time) *string {
	if t == nil {
		return nil
	}
	s := formatTimeText(*t)
	return &s
}

func parseOptionalTimeText(text *string) (*time.Time, error) {
	if text == nil || *text == "" {
		return nil, nil
	}
	t, err := time.Parse(time.RFC3339, *text)
	if err != nil {
		return nil, fmt.Errorf("parsing time %q: %w", *text, err)
	}
	return &t, nil
}

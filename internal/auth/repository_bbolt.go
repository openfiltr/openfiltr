package auth

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/openfiltr/openfiltr/internal/storage"
	bolt "go.etcd.io/bbolt"
	"golang.org/x/crypto/bcrypt"
)

const (
	boltUsersBucket     = "users"
	boltUsersByUsername = "users_by_username"
	boltAPITokensBucket = "api_tokens"
)

type boltRepository struct {
	db *storage.BoltStore
}

var errBoltTokenMatched = errors.New("token matched")

func newBoltRepository(db *storage.BoltStore) authRepository {
	return &boltRepository{db: db}
}

type boltUserRecord struct {
	ID           string    `json:"id"`
	Username     string    `json:"username"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"password_hash"`
	Role         string    `json:"role"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type boltTokenRecord struct {
	ID         string     `json:"id"`
	UserID     string     `json:"user_id"`
	Name       string     `json:"name"`
	TokenHash  string     `json:"token_hash"`
	Scopes     string     `json:"scopes"`
	LastUsedAt *time.Time `json:"last_used_at,omitempty"`
	ExpiresAt  *time.Time `json:"expires_at,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
}

func (r *boltRepository) LookupUserByUsername(username string) (UserRecord, error) {
	var user UserRecord
	err := r.db.View(func(tx *bolt.Tx) error {
		u, err := boltUserByUsernameTx(tx, username)
		if err != nil {
			return err
		}
		user = boltUserToRecord(u)
		return nil
	})
	return user, err
}

func (r *boltRepository) LookupUserByID(id string) (UserRecord, error) {
	var user UserRecord
	err := r.db.View(func(tx *bolt.Tx) error {
		u, err := boltUserByIDTx(tx, id)
		if err != nil {
			return err
		}
		user = boltUserToRecord(u)
		return nil
	})
	return user, err
}

func (r *boltRepository) CountUsers() (int, error) {
	var count int
	err := r.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(boltUsersBucket))
		if b == nil {
			return fmt.Errorf("users bucket missing")
		}
		count = b.Stats().KeyN
		return nil
	})
	return count, err
}

func (r *boltRepository) CreateAdminUser(id, username, email, passwordHash string) error {
	now := time.Now().UTC()
	record := boltUserRecord{
		ID:           id,
		Username:     username,
		Email:        email,
		PasswordHash: passwordHash,
		Role:         "admin",
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	data, err := json.Marshal(record)
	if err != nil {
		return fmt.Errorf("encoding user: %w", err)
	}

	return r.db.Update(func(tx *bolt.Tx) error {
		users := tx.Bucket([]byte(boltUsersBucket))
		if users == nil {
			return fmt.Errorf("users bucket missing")
		}
		if users.Stats().KeyN > 0 {
			return fmt.Errorf("users already exist")
		}
		if err := users.Put([]byte(record.ID), data); err != nil {
			return fmt.Errorf("writing user: %w", err)
		}
		index := tx.Bucket([]byte(boltUsersByUsername))
		if index == nil {
			return fmt.Errorf("users by username bucket missing")
		}
		if err := index.Put([]byte(record.Username), []byte(record.ID)); err != nil {
			return fmt.Errorf("writing username index: %w", err)
		}
		return nil
	})
}

func (r *boltRepository) ListAPITokens(userID string) ([]APITokenView, error) {
	var items []APITokenView
	err := r.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(boltAPITokensBucket))
		if b == nil {
			return fmt.Errorf("api tokens bucket missing")
		}
		return b.ForEach(func(_, v []byte) error {
			var rec boltTokenRecord
			if err := json.Unmarshal(v, &rec); err != nil {
				return nil
			}
			if rec.UserID != userID {
				return nil
			}
			items = append(items, boltTokenToView(rec))
			return nil
		})
	})
	if err != nil {
		return nil, err
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].CreatedAt > items[j].CreatedAt
	})
	return items, nil
}

func (r *boltRepository) CreateAPIToken(id, userID, name, tokenHash string, expiresAt *time.Time) error {
	record := boltTokenRecord{
		ID:        id,
		UserID:    userID,
		Name:      name,
		TokenHash: tokenHash,
		Scopes:    "[]",
		ExpiresAt: expiresAt,
		CreatedAt: time.Now().UTC(),
	}
	data, err := json.Marshal(record)
	if err != nil {
		return fmt.Errorf("encoding api token: %w", err)
	}

	return r.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(boltAPITokensBucket))
		if b == nil {
			return fmt.Errorf("api tokens bucket missing")
		}
		if _, err := boltUserByIDTx(tx, userID); err != nil {
			return err
		}
		if err := b.Put([]byte(record.ID), data); err != nil {
			return fmt.Errorf("writing api token: %w", err)
		}
		return nil
	})
}

func (r *boltRepository) DeleteAPIToken(id, userID string) (bool, error) {
	var deleted bool
	err := r.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(boltAPITokensBucket))
		if b == nil {
			return fmt.Errorf("api tokens bucket missing")
		}
		raw := b.Get([]byte(id))
		if raw == nil {
			return nil
		}
		var rec boltTokenRecord
		if err := json.Unmarshal(raw, &rec); err != nil {
			return fmt.Errorf("decoding api token: %w", err)
		}
		if rec.UserID != userID {
			return nil
		}
		if err := b.Delete([]byte(id)); err != nil {
			return fmt.Errorf("deleting api token: %w", err)
		}
		deleted = true
		return nil
	})
	return deleted, err
}

func (r *boltRepository) ValidateAPIToken(token string) (*Claims, error) {
	type match struct {
		token boltTokenRecord
		user  boltUserRecord
	}
	var found *match

	err := r.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(boltAPITokensBucket))
		if b == nil {
			return fmt.Errorf("api tokens bucket missing")
		}
		return b.ForEach(func(_, v []byte) error {
			var rec boltTokenRecord
			if err := json.Unmarshal(v, &rec); err != nil {
				return nil
			}
			if rec.ExpiresAt != nil && rec.ExpiresAt.Before(time.Now().UTC()) {
				return nil
			}
			if bcrypt.CompareHashAndPassword([]byte(rec.TokenHash), []byte(token)) != nil {
				return nil
			}
			user, err := boltUserByIDTx(tx, rec.UserID)
			if err != nil {
				return err
			}
			found = &match{token: rec, user: user}
			return errBoltTokenMatched
		})
	})
	if err != nil && !errors.Is(err, errBoltTokenMatched) {
		return nil, err
	}
	if found == nil {
		return nil, fmt.Errorf("invalid API token")
	}

	now := time.Now().UTC()
	if err := r.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(boltAPITokensBucket))
		if b == nil {
			return fmt.Errorf("api tokens bucket missing")
		}
		raw := b.Get([]byte(found.token.ID))
		if raw == nil {
			return nil
		}
		var rec boltTokenRecord
		if err := json.Unmarshal(raw, &rec); err != nil {
			return fmt.Errorf("decoding api token: %w", err)
		}
		rec.LastUsedAt = &now
		data, err := json.Marshal(rec)
		if err != nil {
			return fmt.Errorf("encoding api token: %w", err)
		}
		if err := b.Put([]byte(rec.ID), data); err != nil {
			return fmt.Errorf("updating api token: %w", err)
		}
		return nil
	}); err != nil {
		return nil, err
	}

	return &Claims{UserID: found.user.ID, Username: found.user.Username, Role: found.user.Role}, nil
}

func boltUserByUsernameTx(tx *bolt.Tx, username string) (boltUserRecord, error) {
	index := tx.Bucket([]byte(boltUsersByUsername))
	if index == nil {
		return boltUserRecord{}, fmt.Errorf("users by username bucket missing")
	}
	id := index.Get([]byte(username))
	if id == nil {
		return boltUserRecord{}, sql.ErrNoRows
	}
	return boltUserByIDTx(tx, string(id))
}

func boltUserByIDTx(tx *bolt.Tx, id string) (boltUserRecord, error) {
	users := tx.Bucket([]byte(boltUsersBucket))
	if users == nil {
		return boltUserRecord{}, fmt.Errorf("users bucket missing")
	}
	raw := users.Get([]byte(id))
	if raw == nil {
		return boltUserRecord{}, sql.ErrNoRows
	}
	var user boltUserRecord
	if err := json.Unmarshal(raw, &user); err != nil {
		return boltUserRecord{}, fmt.Errorf("decoding user: %w", err)
	}
	return user, nil
}

func boltUserToRecord(user boltUserRecord) UserRecord {
	return UserRecord{
		ID:           user.ID,
		Username:     user.Username,
		PasswordHash: user.PasswordHash,
		Role:         user.Role,
	}
}

func boltTokenToView(token boltTokenRecord) APITokenView {
	return APITokenView{
		ID:         token.ID,
		Name:       token.Name,
		Scopes:     token.Scopes,
		LastUsedAt: formatOptionalTimeText(token.LastUsedAt),
		ExpiresAt:  formatOptionalTimeText(token.ExpiresAt),
		CreatedAt:  formatTimeText(token.CreatedAt),
	}
}

var _ authRepository = (*boltRepository)(nil)

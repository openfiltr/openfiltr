package auth

import (
	"fmt"
	"time"

	"github.com/openfiltr/openfiltr/internal/storage"
	"golang.org/x/crypto/bcrypt"
)

type sqlRepository struct {
	db storage.Store
}

func newSQLRepository(db storage.Store) authRepository {
	return &sqlRepository{db: db}
}

func (r *sqlRepository) LookupUserByUsername(username string) (UserRecord, error) {
	var user UserRecord
	user.Username = username
	if err := r.db.QueryRow(storage.Rebind("SELECT id,password_hash,role FROM users WHERE username=?"), username).Scan(&user.ID, &user.PasswordHash, &user.Role); err != nil {
		return UserRecord{}, err
	}
	return user, nil
}

func (r *sqlRepository) LookupUserByID(id string) (UserRecord, error) {
	var user UserRecord
	user.ID = id
	if err := r.db.QueryRow(storage.Rebind("SELECT username,password_hash,role FROM users WHERE id=?"), id).Scan(&user.Username, &user.PasswordHash, &user.Role); err != nil {
		return UserRecord{}, err
	}
	return user, nil
}

func (r *sqlRepository) CountUsers() (int, error) {
	var count int
	if err := r.db.QueryRow("SELECT COUNT(*) FROM users").Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

func (r *sqlRepository) CreateAdminUser(id, username, email, passwordHash string) error {
	_, err := r.db.Exec(storage.Rebind(`INSERT INTO users(id,username,email,password_hash,role) VALUES(?,?,?,?,'admin')`), id, username, email, passwordHash)
	return err
}

func (r *sqlRepository) ListAPITokens(userID string) ([]APITokenView, error) {
	rows, err := r.db.Query(storage.Rebind(`SELECT id,name,scopes,last_used_at::text,expires_at::text,created_at::text FROM api_tokens WHERE user_id=? ORDER BY created_at DESC`), userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []APITokenView
	for rows.Next() {
		var it APITokenView
		if err := rows.Scan(&it.ID, &it.Name, &it.Scopes, &it.LastUsedAt, &it.ExpiresAt, &it.CreatedAt); err != nil {
			continue
		}
		items = append(items, it)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func (r *sqlRepository) CreateAPIToken(id, userID, name, tokenHash string, expiresAt *time.Time) error {
	var exp interface{}
	if expiresAt != nil {
		exp = expiresAt.UTC().Format("2006-01-02 15:04:05")
	}
	_, err := r.db.Exec(storage.Rebind(`INSERT INTO api_tokens(id,user_id,name,token_hash,scopes,expires_at) VALUES(?,?,?,?,'[]',?)`), id, userID, name, tokenHash, exp)
	return err
}

func (r *sqlRepository) DeleteAPIToken(id, userID string) (bool, error) {
	res, err := r.db.Exec(storage.Rebind("DELETE FROM api_tokens WHERE id=? AND user_id=?"), id, userID)
	if err != nil {
		return false, err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return false, err
	}
	return n > 0, nil
}

func (r *sqlRepository) ValidateAPIToken(token string) (*Claims, error) {
	rows, err := r.db.Query(`SELECT at.token_hash,u.id,u.username,u.role FROM api_tokens at JOIN users u ON u.id=at.user_id WHERE (at.expires_at IS NULL OR at.expires_at>NOW())`)
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
			_, _ = r.db.Exec(storage.Rebind("UPDATE api_tokens SET last_used_at=NOW() WHERE token_hash=?"), h)
			return &Claims{UserID: uid, Username: uname, Role: role}, nil
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return nil, fmt.Errorf("invalid API token")
}

var _ authRepository = (*sqlRepository)(nil)

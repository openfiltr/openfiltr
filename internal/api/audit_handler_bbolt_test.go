package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"github.com/openfiltr/openfiltr/internal/storage"
	bolt "go.etcd.io/bbolt"
)

func TestBoltListAuditEvents(t *testing.T) {
	store := openBoltAuditHandlerStore(t)
	defer store.Close()

	createdAt := time.Now().UTC().Format(time.RFC3339)
	event := storage.AuditEventView{
		ID:           "audit-1",
		Action:       "create",
		ResourceType: "user",
		CreatedAt:    createdAt,
	}
	data, err := json.Marshal(event)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}
	if err := store.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("audit_events"))
		if b == nil {
			return fmt.Errorf("audit_events bucket missing")
		}
		return b.Put([]byte(event.ID), data)
	}); err != nil {
		t.Fatalf("writing audit event = %v", err)
	}

	handler := &Handler{db: store}
	rec := httptest.NewRecorder()
	handler.ListAuditEvents(rec, httptest.NewRequest(http.MethodGet, "/api/v1/audit?limit=10", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("ListAuditEvents() status = %d, want %d", rec.Code, http.StatusOK)
	}
	var payload struct {
		Total int          `json:"total"`
		Items []auditEvent `json:"items"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&payload); err != nil {
		t.Fatalf("decoding ListAuditEvents() response: %v", err)
	}
	if payload.Total != 1 {
		t.Fatalf("ListAuditEvents() total = %d, want %d", payload.Total, 1)
	}
	if len(payload.Items) != 1 {
		t.Fatalf("ListAuditEvents() items len = %d, want %d", len(payload.Items), 1)
	}
	if payload.Items[0].ID != event.ID || payload.Items[0].Action != event.Action {
		t.Fatalf("ListAuditEvents() item = %#v, want %#v", payload.Items[0], event)
	}
}

func openBoltAuditHandlerStore(t *testing.T) *storage.BoltStore {
	t.Helper()
	path := filepath.Join(t.TempDir(), "openfiltr.db")
	store, err := storage.OpenBolt(path)
	if err != nil {
		t.Fatalf("OpenBolt() error = %v", err)
	}
	return store
}

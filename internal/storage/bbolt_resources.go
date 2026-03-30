package storage

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"sort"

	bolt "go.etcd.io/bbolt"
)

const (
	clientsBucket         = "clients"
	groupsBucket          = "groups"
	upstreamServersBucket = "upstream_servers"
	ruleSourcesBucket     = "rule_sources"
	auditEventsBucket     = "audit_events"
)

func (s *BoltStore) ListClients(limit, offset int) ([]ClientView, int, error) {
	items, total, err := listBucketViews[ClientView](s, clientsBucket)
	if err != nil {
		return nil, 0, err
	}
	sort.Slice(items, func(i, j int) bool { return items[i].CreatedAt > items[j].CreatedAt })
	return paginateClientViews(items, limit, offset), total, nil
}

func (s *BoltStore) GetClient(id string) (ClientView, error) {
	return getBucketView[ClientView](s, clientsBucket, id)
}

func (s *BoltStore) CreateClient(id, name, identifier, identifierType string, groupID, comment *string) (ClientView, error) {
	if groupID != nil && *groupID != "" {
		if _, err := s.GetGroup(*groupID); err != nil {
			return ClientView{}, err
		}
	}
	rec := ClientView{
		ID:             id,
		Name:           name,
		Identifier:     identifier,
		IdentifierType: identifierType,
		GroupID:        groupID,
		Comment:        comment,
		CreatedAt:      utcNowText(),
		UpdatedAt:      utcNowText(),
	}
	if err := putBucketView(s, clientsBucket, rec.ID, rec); err != nil {
		return ClientView{}, err
	}
	return rec, nil
}

func (s *BoltStore) UpdateClient(id string, name, identifier, identifierType, groupID, comment *string) (ClientView, bool, error) {
	var updated ClientView
	var found bool
	err := s.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(clientsBucket))
		if b == nil {
			return fmt.Errorf("%s bucket missing", clientsBucket)
		}
		raw := b.Get([]byte(id))
		if raw == nil {
			return nil
		}
		var current ClientView
		if err := json.Unmarshal(raw, &current); err != nil {
			return fmt.Errorf("decoding client: %w", err)
		}
		if groupID != nil && *groupID != "" {
			if _, err := getGroupTx(tx, *groupID); err != nil {
				return err
			}
		}
		updated = current
		if name != nil {
			updated.Name = *name
		}
		if identifier != nil {
			updated.Identifier = *identifier
		}
		if identifierType != nil {
			updated.IdentifierType = *identifierType
		}
		if groupID != nil {
			updated.GroupID = groupID
		}
		if comment != nil {
			updated.Comment = comment
		}
		updated.UpdatedAt = utcNowText()
		data, err := json.Marshal(updated)
		if err != nil {
			return fmt.Errorf("encoding client: %w", err)
		}
		if err := b.Put([]byte(id), data); err != nil {
			return fmt.Errorf("updating client: %w", err)
		}
		found = true
		return nil
	})
	return updated, found, err
}

func (s *BoltStore) DeleteClient(id string) (bool, error) {
	return deleteBucketView(s, clientsBucket, id)
}

func (s *BoltStore) HasClientsForGroup(groupID string) (bool, error) {
	var found bool
	err := s.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(clientsBucket))
		if b == nil {
			return fmt.Errorf("%s bucket missing", clientsBucket)
		}
		return b.ForEach(func(_, v []byte) error {
			var rec ClientView
			if err := json.Unmarshal(v, &rec); err != nil {
				return nil
			}
			if rec.GroupID != nil && *rec.GroupID == groupID {
				found = true
				return fmt.Errorf("group referenced")
			}
			return nil
		})
	})
	if err != nil && err.Error() != "group referenced" {
		return false, err
	}
	return found, nil
}

func (s *BoltStore) ListGroups(limit, offset int) ([]GroupView, int, error) {
	items, total, err := listBucketViews[GroupView](s, groupsBucket)
	if err != nil {
		return nil, 0, err
	}
	sort.Slice(items, func(i, j int) bool { return items[i].Name < items[j].Name })
	return paginateGroupViews(items, limit, offset), total, nil
}

func (s *BoltStore) GetGroup(id string) (GroupView, error) {
	return getBucketView[GroupView](s, groupsBucket, id)
}

func (s *BoltStore) CreateGroup(id, name string, description *string) (GroupView, error) {
	rec := GroupView{
		ID:          id,
		Name:        name,
		Description: description,
		CreatedAt:   utcNowText(),
		UpdatedAt:   utcNowText(),
	}
	if err := putBucketView(s, groupsBucket, rec.ID, rec); err != nil {
		return GroupView{}, err
	}
	return rec, nil
}

func (s *BoltStore) UpdateGroup(id string, name, description *string) (GroupView, bool, error) {
	var updated GroupView
	var found bool
	err := s.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(groupsBucket))
		if b == nil {
			return fmt.Errorf("%s bucket missing", groupsBucket)
		}
		raw := b.Get([]byte(id))
		if raw == nil {
			return nil
		}
		var current GroupView
		if err := json.Unmarshal(raw, &current); err != nil {
			return fmt.Errorf("decoding group: %w", err)
		}
		updated = current
		if name != nil {
			updated.Name = *name
		}
		if description != nil {
			updated.Description = description
		}
		updated.UpdatedAt = utcNowText()
		data, err := json.Marshal(updated)
		if err != nil {
			return fmt.Errorf("encoding group: %w", err)
		}
		if err := b.Put([]byte(id), data); err != nil {
			return fmt.Errorf("updating group: %w", err)
		}
		found = true
		return nil
	})
	return updated, found, err
}

func (s *BoltStore) DeleteGroup(id string) (bool, error) {
	used, err := s.HasClientsForGroup(id)
	if err != nil {
		return false, err
	}
	if used {
		return false, fmt.Errorf("group is referenced by clients")
	}
	return deleteBucketView(s, groupsBucket, id)
}

func (s *BoltStore) ListUpstreamServers(limit, offset int) ([]UpstreamServerView, int, error) {
	items, total, err := listBucketViews[UpstreamServerView](s, upstreamServersBucket)
	if err != nil {
		return nil, 0, err
	}
	sort.Slice(items, func(i, j int) bool {
		if items[i].Priority == items[j].Priority {
			return items[i].CreatedAt > items[j].CreatedAt
		}
		return items[i].Priority < items[j].Priority
	})
	return paginateUpstreamViews(items, limit, offset), total, nil
}

func (s *BoltStore) GetUpstreamServer(id string) (UpstreamServerView, error) {
	return getBucketView[UpstreamServerView](s, upstreamServersBucket, id)
}

func (s *BoltStore) CreateUpstreamServer(id, name, address, protocol string, enabled, priority int) (UpstreamServerView, error) {
	rec := UpstreamServerView{
		ID:        id,
		Name:      name,
		Address:   address,
		Protocol:  protocol,
		Enabled:   enabled,
		Priority:  priority,
		CreatedAt: utcNowText(),
		UpdatedAt: utcNowText(),
	}
	if err := putBucketView(s, upstreamServersBucket, rec.ID, rec); err != nil {
		return UpstreamServerView{}, err
	}
	return rec, nil
}

func (s *BoltStore) UpdateUpstreamServer(id string, name, address, protocol *string, enabled, priority *int) (UpstreamServerView, bool, error) {
	var updated UpstreamServerView
	var found bool
	err := s.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(upstreamServersBucket))
		if b == nil {
			return fmt.Errorf("%s bucket missing", upstreamServersBucket)
		}
		raw := b.Get([]byte(id))
		if raw == nil {
			return nil
		}
		var current UpstreamServerView
		if err := json.Unmarshal(raw, &current); err != nil {
			return fmt.Errorf("decoding upstream server: %w", err)
		}
		updated = current
		if name != nil {
			updated.Name = *name
		}
		if address != nil {
			updated.Address = *address
		}
		if protocol != nil {
			updated.Protocol = *protocol
		}
		if enabled != nil {
			updated.Enabled = *enabled
		}
		if priority != nil {
			updated.Priority = *priority
		}
		updated.UpdatedAt = utcNowText()
		data, err := json.Marshal(updated)
		if err != nil {
			return fmt.Errorf("encoding upstream server: %w", err)
		}
		if err := b.Put([]byte(id), data); err != nil {
			return fmt.Errorf("updating upstream server: %w", err)
		}
		found = true
		return nil
	})
	return updated, found, err
}

func (s *BoltStore) DeleteUpstreamServer(id string) (bool, error) {
	return deleteBucketView(s, upstreamServersBucket, id)
}

func (s *BoltStore) ListRuleSources(limit, offset int) ([]RuleSourceView, int, error) {
	items, total, err := listBucketViews[RuleSourceView](s, ruleSourcesBucket)
	if err != nil {
		return nil, 0, err
	}
	sort.Slice(items, func(i, j int) bool { return items[i].CreatedAt > items[j].CreatedAt })
	return paginateRuleSourceViews(items, limit, offset), total, nil
}

func (s *BoltStore) GetRuleSource(id string) (RuleSourceView, error) {
	return getBucketView[RuleSourceView](s, ruleSourcesBucket, id)
}

func (s *BoltStore) CreateRuleSource(id, name, url, format string, enabled int) (RuleSourceView, error) {
	rec := RuleSourceView{
		ID:        id,
		Name:      name,
		URL:       url,
		Format:    format,
		Enabled:   enabled,
		CreatedAt: utcNowText(),
		UpdatedAt: utcNowText(),
	}
	if err := putBucketView(s, ruleSourcesBucket, rec.ID, rec); err != nil {
		return RuleSourceView{}, err
	}
	return rec, nil
}

func (s *BoltStore) UpdateRuleSource(id string, name, url, format *string, enabled *int) (RuleSourceView, bool, error) {
	var updated RuleSourceView
	var found bool
	err := s.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(ruleSourcesBucket))
		if b == nil {
			return fmt.Errorf("%s bucket missing", ruleSourcesBucket)
		}
		raw := b.Get([]byte(id))
		if raw == nil {
			return nil
		}
		var current RuleSourceView
		if err := json.Unmarshal(raw, &current); err != nil {
			return fmt.Errorf("decoding rule source: %w", err)
		}
		updated = current
		if name != nil {
			updated.Name = *name
		}
		if url != nil {
			updated.URL = *url
		}
		if format != nil {
			updated.Format = *format
		}
		if enabled != nil {
			updated.Enabled = *enabled
		}
		updated.UpdatedAt = utcNowText()
		data, err := json.Marshal(updated)
		if err != nil {
			return fmt.Errorf("encoding rule source: %w", err)
		}
		if err := b.Put([]byte(id), data); err != nil {
			return fmt.Errorf("updating rule source: %w", err)
		}
		found = true
		return nil
	})
	return updated, found, err
}

func (s *BoltStore) DeleteRuleSource(id string) (bool, error) {
	return deleteBucketView(s, ruleSourcesBucket, id)
}

func (s *BoltStore) RefreshRuleSource(id string) (bool, error) {
	var found bool
	err := s.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(ruleSourcesBucket))
		if b == nil {
			return fmt.Errorf("%s bucket missing", ruleSourcesBucket)
		}
		raw := b.Get([]byte(id))
		if raw == nil {
			return nil
		}
		var current RuleSourceView
		if err := json.Unmarshal(raw, &current); err != nil {
			return fmt.Errorf("decoding rule source: %w", err)
		}
		now := utcNowText()
		current.LastUpdatedAt = &now
		current.UpdatedAt = now
		data, err := json.Marshal(current)
		if err != nil {
			return fmt.Errorf("encoding rule source: %w", err)
		}
		if err := b.Put([]byte(id), data); err != nil {
			return fmt.Errorf("refreshing rule source: %w", err)
		}
		found = true
		return nil
	})
	return found, err
}

func (s *BoltStore) ListAuditEvents(limit, offset int) ([]AuditEventView, int, error) {
	items, total, err := listBucketViews[AuditEventView](s, auditEventsBucket)
	if err != nil {
		return nil, 0, err
	}
	sort.Slice(items, func(i, j int) bool { return items[i].CreatedAt > items[j].CreatedAt })
	return paginateAuditViews(items, limit, offset), total, nil
}

func listBucketViews[T any](s *BoltStore, bucketName string) ([]T, int, error) {
	items := make([]T, 0)
	total := 0
	err := s.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucketName))
		if b == nil {
			return fmt.Errorf("%s bucket missing", bucketName)
		}
		total = b.Stats().KeyN
		return b.ForEach(func(_, v []byte) error {
			var rec T
			if err := json.Unmarshal(v, &rec); err != nil {
				return nil
			}
			items = append(items, rec)
			return nil
		})
	})
	return items, total, err
}

func getBucketView[T any](s *BoltStore, bucketName, id string) (T, error) {
	var rec T
	err := s.View(func(tx *bolt.Tx) error {
		found, err := getBucketViewTx[T](tx, bucketName, id)
		if err != nil {
			return err
		}
		rec = found
		return nil
	})
	return rec, err
}

func getBucketViewTx[T any](tx *bolt.Tx, bucketName, id string) (T, error) {
	var rec T
	b := tx.Bucket([]byte(bucketName))
	if b == nil {
		return rec, fmt.Errorf("%s bucket missing", bucketName)
	}
	raw := b.Get([]byte(id))
	if raw == nil {
		return rec, sql.ErrNoRows
	}
	if err := json.Unmarshal(raw, &rec); err != nil {
		return rec, fmt.Errorf("decoding %s: %w", bucketName, err)
	}
	return rec, nil
}

func getGroupTx(tx *bolt.Tx, id string) (GroupView, error) {
	return getBucketViewTx[GroupView](tx, groupsBucket, id)
}

func putBucketView[T any](s *BoltStore, bucketName, id string, rec T) error {
	data, err := json.Marshal(rec)
	if err != nil {
		return fmt.Errorf("encoding %s: %w", bucketName, err)
	}
	return s.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucketName))
		if b == nil {
			return fmt.Errorf("%s bucket missing", bucketName)
		}
		if err := b.Put([]byte(id), data); err != nil {
			return fmt.Errorf("writing %s: %w", bucketName, err)
		}
		return nil
	})
}

func deleteBucketView(s *BoltStore, bucketName, id string) (bool, error) {
	var deleted bool
	err := s.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucketName))
		if b == nil {
			return fmt.Errorf("%s bucket missing", bucketName)
		}
		raw := b.Get([]byte(id))
		if raw == nil {
			return nil
		}
		if err := b.Delete([]byte(id)); err != nil {
			return fmt.Errorf("deleting %s: %w", bucketName, err)
		}
		deleted = true
		return nil
	})
	return deleted, err
}

func paginateClientViews(items []ClientView, limit, offset int) []ClientView {
	if offset >= len(items) {
		return []ClientView{}
	}
	end := offset + limit
	if end > len(items) {
		end = len(items)
	}
	return items[offset:end]
}

func paginateGroupViews(items []GroupView, limit, offset int) []GroupView {
	if offset >= len(items) {
		return []GroupView{}
	}
	end := offset + limit
	if end > len(items) {
		end = len(items)
	}
	return items[offset:end]
}

func paginateUpstreamViews(items []UpstreamServerView, limit, offset int) []UpstreamServerView {
	if offset >= len(items) {
		return []UpstreamServerView{}
	}
	end := offset + limit
	if end > len(items) {
		end = len(items)
	}
	return items[offset:end]
}

func paginateRuleSourceViews(items []RuleSourceView, limit, offset int) []RuleSourceView {
	if offset >= len(items) {
		return []RuleSourceView{}
	}
	end := offset + limit
	if end > len(items) {
		end = len(items)
	}
	return items[offset:end]
}

func paginateAuditViews(items []AuditEventView, limit, offset int) []AuditEventView {
	if offset >= len(items) {
		return []AuditEventView{}
	}
	end := offset + limit
	if end > len(items) {
		end = len(items)
	}
	return items[offset:end]
}

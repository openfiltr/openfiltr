package storage

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	bolt "go.etcd.io/bbolt"
)

func (s *BoltStore) ListRules(table string, limit, offset int) ([]RuleView, int, error) {
	items := make([]RuleView, 0)
	total := 0
	if err := s.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(table))
		if b == nil {
			return fmt.Errorf("%s bucket missing", table)
		}
		total = b.Stats().KeyN
		return b.ForEach(func(_, v []byte) error {
			var rec RuleView
			if err := json.Unmarshal(v, &rec); err != nil {
				return nil
			}
			items = append(items, rec)
			return nil
		})
	}); err != nil {
		return nil, 0, err
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].CreatedAt > items[j].CreatedAt
	})
	return paginateRuleViews(items, limit, offset), total, nil
}

func (s *BoltStore) GetRule(table, id string) (RuleView, error) {
	var rec RuleView
	err := s.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(table))
		if b == nil {
			return fmt.Errorf("%s bucket missing", table)
		}
		raw := b.Get([]byte(id))
		if raw == nil {
			return sql.ErrNoRows
		}
		if err := json.Unmarshal(raw, &rec); err != nil {
			return fmt.Errorf("decoding %s: %w", table, err)
		}
		return nil
	})
	return rec, err
}

func (s *BoltStore) CreateRule(table, id, pattern, ruleType string, comment *string, enabled int, createdBy *string) (RuleView, error) {
	now := utcNowText()
	rec := RuleView{
		ID:        id,
		Pattern:   pattern,
		RuleType:  ruleType,
		Comment:   comment,
		Enabled:   enabled,
		CreatedBy: createdBy,
		CreatedAt: now,
		UpdatedAt: now,
	}
	data, err := json.Marshal(rec)
	if err != nil {
		return RuleView{}, fmt.Errorf("encoding %s: %w", table, err)
	}

	if err := s.Update(func(tx *bolt.Tx) error {
		main, lookup, err := ruleBuckets(tx, table)
		if err != nil {
			return err
		}
		if err := main.Put([]byte(rec.ID), data); err != nil {
			return fmt.Errorf("writing %s: %w", table, err)
		}
		if rec.Enabled == 1 {
			if err := putRuleLookup(lookup, rec.Pattern, rec.RuleType, rec.ID); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return RuleView{}, err
	}
	return rec, nil
}

func (s *BoltStore) UpdateRule(table, id string, pattern, ruleType, comment *string, enabled *int) (RuleView, bool, error) {
	var updated RuleView
	var found bool
	err := s.Update(func(tx *bolt.Tx) error {
		main, lookup, err := ruleBuckets(tx, table)
		if err != nil {
			return err
		}
		raw := main.Get([]byte(id))
		if raw == nil {
			return nil
		}
		var current RuleView
		if err := json.Unmarshal(raw, &current); err != nil {
			return fmt.Errorf("decoding %s: %w", table, err)
		}

		updated = current
		if pattern != nil {
			updated.Pattern = *pattern
		}
		if ruleType != nil {
			updated.RuleType = *ruleType
		}
		if comment != nil {
			updated.Comment = comment
		}
		if enabled != nil {
			updated.Enabled = *enabled
		}
		updated.UpdatedAt = utcNowText()

		if current.Enabled == 1 {
			if err := deleteRuleLookup(lookup, current.Pattern, current.RuleType, current.ID); err != nil {
				return err
			}
		}
		if updated.Enabled == 1 {
			if err := putRuleLookup(lookup, updated.Pattern, updated.RuleType, updated.ID); err != nil {
				return err
			}
		}

		data, err := json.Marshal(updated)
		if err != nil {
			return fmt.Errorf("encoding %s: %w", table, err)
		}
		if err := main.Put([]byte(id), data); err != nil {
			return fmt.Errorf("updating %s: %w", table, err)
		}
		found = true
		return nil
	})
	return updated, found, err
}

func (s *BoltStore) DeleteRule(table, id string) (bool, error) {
	var deleted bool
	err := s.Update(func(tx *bolt.Tx) error {
		main, lookup, err := ruleBuckets(tx, table)
		if err != nil {
			return err
		}
		raw := main.Get([]byte(id))
		if raw == nil {
			return nil
		}
		var rec RuleView
		if err := json.Unmarshal(raw, &rec); err != nil {
			return fmt.Errorf("decoding %s: %w", table, err)
		}
		if rec.Enabled == 1 {
			if err := deleteRuleLookup(lookup, rec.Pattern, rec.RuleType, rec.ID); err != nil {
				return err
			}
		}
		if err := main.Delete([]byte(id)); err != nil {
			return fmt.Errorf("deleting %s: %w", table, err)
		}
		deleted = true
		return nil
	})
	return deleted, err
}

func (s *BoltStore) HasRulePattern(table, ruleType, pattern string) (bool, error) {
	var exists bool
	err := s.View(func(tx *bolt.Tx) error {
		_, lookup, err := ruleBuckets(tx, table)
		if err != nil {
			return err
		}
		prefix := ruleLookupPrefix(pattern, ruleType)
		c := lookup.Cursor()
		k, _ := c.Seek(prefix)
		if k != nil && bytes.HasPrefix(k, prefix) {
			exists = true
		}
		return nil
	})
	return exists, err
}

func (s *BoltStore) ListRulePatternsByType(table, ruleType string) ([]string, error) {
	patterns := make([]string, 0)
	err := s.View(func(tx *bolt.Tx) error {
		main := tx.Bucket([]byte(table))
		if main == nil {
			return fmt.Errorf("%s bucket missing", table)
		}
		return main.ForEach(func(_, v []byte) error {
			var rec RuleView
			if err := json.Unmarshal(v, &rec); err != nil {
				return nil
			}
			if rec.Enabled != 1 || rec.RuleType != ruleType {
				return nil
			}
			patterns = append(patterns, rec.Pattern)
			return nil
		})
	})
	if err != nil {
		return nil, err
	}
	sort.Strings(patterns)
	return patterns, nil
}

func (s *BoltStore) ListDNSEntries(limit, offset int) ([]DNSEntryView, int, error) {
	items := make([]DNSEntryView, 0)
	total := 0
	if err := s.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(dnsEntriesBucket))
		if b == nil {
			return fmt.Errorf("%s bucket missing", dnsEntriesBucket)
		}
		total = b.Stats().KeyN
		return b.ForEach(func(_, v []byte) error {
			var rec DNSEntryView
			if err := json.Unmarshal(v, &rec); err != nil {
				return nil
			}
			items = append(items, rec)
			return nil
		})
	}); err != nil {
		return nil, 0, err
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].CreatedAt > items[j].CreatedAt
	})
	return paginateDNSEntryViews(items, limit, offset), total, nil
}

func (s *BoltStore) GetDNSEntry(id string) (DNSEntryView, error) {
	var rec DNSEntryView
	err := s.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(dnsEntriesBucket))
		if b == nil {
			return fmt.Errorf("%s bucket missing", dnsEntriesBucket)
		}
		raw := b.Get([]byte(id))
		if raw == nil {
			return sql.ErrNoRows
		}
		if err := json.Unmarshal(raw, &rec); err != nil {
			return fmt.Errorf("decoding %s: %w", dnsEntriesBucket, err)
		}
		return nil
	})
	return rec, err
}

func (s *BoltStore) CreateDNSEntry(id, host, entryType, value string, ttl int, comment, createdBy *string, enabled int) (DNSEntryView, error) {
	now := utcNowText()
	rec := DNSEntryView{
		ID:        id,
		Host:      host,
		EntryType: entryType,
		Value:     value,
		TTL:       ttl,
		Comment:   comment,
		Enabled:   enabled,
		CreatedBy: createdBy,
		CreatedAt: now,
		UpdatedAt: now,
	}
	data, err := json.Marshal(rec)
	if err != nil {
		return DNSEntryView{}, fmt.Errorf("encoding dns entry: %w", err)
	}

	if err := s.Update(func(tx *bolt.Tx) error {
		main, lookup, err := dnsEntryBuckets(tx)
		if err != nil {
			return err
		}
		if err := main.Put([]byte(rec.ID), data); err != nil {
			return fmt.Errorf("writing dns entry: %w", err)
		}
		if rec.Enabled == 1 {
			if err := putDNSLookup(lookup, rec.Host, rec.EntryType, rec.ID); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return DNSEntryView{}, err
	}
	return rec, nil
}

func (s *BoltStore) UpdateDNSEntry(id string, host, entryType, value *string, ttl *int, comment *string, enabled *int) (DNSEntryView, bool, error) {
	var updated DNSEntryView
	var found bool
	err := s.Update(func(tx *bolt.Tx) error {
		main, lookup, err := dnsEntryBuckets(tx)
		if err != nil {
			return err
		}
		raw := main.Get([]byte(id))
		if raw == nil {
			return nil
		}
		var current DNSEntryView
		if err := json.Unmarshal(raw, &current); err != nil {
			return fmt.Errorf("decoding dns entry: %w", err)
		}

		updated = current
		if host != nil {
			updated.Host = *host
		}
		if entryType != nil {
			updated.EntryType = *entryType
		}
		if value != nil {
			updated.Value = *value
		}
		if ttl != nil {
			updated.TTL = *ttl
		}
		if comment != nil {
			updated.Comment = comment
		}
		if enabled != nil {
			updated.Enabled = *enabled
		}
		updated.UpdatedAt = utcNowText()

		if current.Enabled == 1 {
			if err := deleteDNSLookup(lookup, current.Host, current.EntryType, current.ID); err != nil {
				return err
			}
		}
		if updated.Enabled == 1 {
			if err := putDNSLookup(lookup, updated.Host, updated.EntryType, updated.ID); err != nil {
				return err
			}
		}

		data, err := json.Marshal(updated)
		if err != nil {
			return fmt.Errorf("encoding dns entry: %w", err)
		}
		if err := main.Put([]byte(id), data); err != nil {
			return fmt.Errorf("updating dns entry: %w", err)
		}
		found = true
		return nil
	})
	return updated, found, err
}

func (s *BoltStore) DeleteDNSEntry(id string) (bool, error) {
	var deleted bool
	err := s.Update(func(tx *bolt.Tx) error {
		main, lookup, err := dnsEntryBuckets(tx)
		if err != nil {
			return err
		}
		raw := main.Get([]byte(id))
		if raw == nil {
			return nil
		}
		var rec DNSEntryView
		if err := json.Unmarshal(raw, &rec); err != nil {
			return fmt.Errorf("decoding dns entry: %w", err)
		}
		if rec.Enabled == 1 {
			if err := deleteDNSLookup(lookup, rec.Host, rec.EntryType, rec.ID); err != nil {
				return err
			}
		}
		if err := main.Delete([]byte(id)); err != nil {
			return fmt.Errorf("deleting dns entry: %w", err)
		}
		deleted = true
		return nil
	})
	return deleted, err
}

func (s *BoltStore) DNSEntriesByHostAndType(host, entryType string) ([]DNSEntryView, error) {
	items := make([]DNSEntryView, 0)
	err := s.View(func(tx *bolt.Tx) error {
		main, lookup, err := dnsEntryBuckets(tx)
		if err != nil {
			return err
		}
		prefix := dnsLookupPrefix(host, entryType)
		cursor := lookup.Cursor()
		for k, _ := cursor.Seek(prefix); k != nil && bytes.HasPrefix(k, prefix); k, _ = cursor.Next() {
			parts := strings.SplitN(string(k), "\x00", 3)
			if len(parts) != 3 {
				continue
			}
			raw := main.Get([]byte(parts[2]))
			if raw == nil {
				continue
			}
			var rec DNSEntryView
			if err := json.Unmarshal(raw, &rec); err != nil {
				continue
			}
			items = append(items, rec)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].CreatedAt > items[j].CreatedAt
	})
	return items, nil
}

const (
	dnsEntriesBucket       = "dns_entries"
	dnsEntriesLookupBucket = "dns_entries_lookup"
)

func ruleBuckets(tx *bolt.Tx, table string) (*bolt.Bucket, *bolt.Bucket, error) {
	main := tx.Bucket([]byte(table))
	if main == nil {
		return nil, nil, fmt.Errorf("%s bucket missing", table)
	}
	lookup := tx.Bucket([]byte(ruleLookupBucketName(table)))
	if lookup == nil {
		return nil, nil, fmt.Errorf("%s bucket missing", ruleLookupBucketName(table))
	}
	return main, lookup, nil
}

func dnsEntryBuckets(tx *bolt.Tx) (*bolt.Bucket, *bolt.Bucket, error) {
	main := tx.Bucket([]byte(dnsEntriesBucket))
	if main == nil {
		return nil, nil, fmt.Errorf("%s bucket missing", dnsEntriesBucket)
	}
	lookup := tx.Bucket([]byte(dnsEntriesLookupBucket))
	if lookup == nil {
		return nil, nil, fmt.Errorf("%s bucket missing", dnsEntriesLookupBucket)
	}
	return main, lookup, nil
}

func ruleLookupBucketName(table string) string {
	return table + "_lookup"
}

func ruleLookupPrefix(pattern, ruleType string) []byte {
	return []byte(normaliseLookupKey(pattern) + "\x00" + ruleType + "\x00")
}

func dnsLookupPrefix(host, entryType string) []byte {
	return []byte(normaliseLookupKey(host) + "\x00" + entryType + "\x00")
}

func putRuleLookup(bucket *bolt.Bucket, pattern, ruleType, id string) error {
	key := append(ruleLookupPrefix(pattern, ruleType), []byte(id)...)
	return bucket.Put(key, nil)
}

func deleteRuleLookup(bucket *bolt.Bucket, pattern, ruleType, id string) error {
	key := append(ruleLookupPrefix(pattern, ruleType), []byte(id)...)
	return bucket.Delete(key)
}

func putDNSLookup(bucket *bolt.Bucket, host, entryType, id string) error {
	key := append(dnsLookupPrefix(host, entryType), []byte(id)...)
	return bucket.Put(key, nil)
}

func deleteDNSLookup(bucket *bolt.Bucket, host, entryType, id string) error {
	key := append(dnsLookupPrefix(host, entryType), []byte(id)...)
	return bucket.Delete(key)
}

func normaliseLookupKey(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}

func paginateRuleViews(items []RuleView, limit, offset int) []RuleView {
	if offset >= len(items) {
		return []RuleView{}
	}
	end := offset + limit
	if end > len(items) {
		end = len(items)
	}
	return items[offset:end]
}

func paginateDNSEntryViews(items []DNSEntryView, limit, offset int) []DNSEntryView {
	if offset >= len(items) {
		return []DNSEntryView{}
	}
	end := offset + limit
	if end > len(items) {
		end = len(items)
	}
	return items[offset:end]
}

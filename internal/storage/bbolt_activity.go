package storage

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	bolt "go.etcd.io/bbolt"
)

func (s *BoltStore) AppendActivityEntry(id, clientIP, domain, queryType, action string, ruleID, ruleSource *string, responseTimeMs *int) error {
	rec := ActivityEntryView{
		ID:             id,
		ClientIP:       clientIP,
		Domain:         domain,
		QueryType:      queryType,
		Action:         action,
		RuleID:         ruleID,
		RuleSource:     ruleSource,
		ResponseTimeMs: responseTimeMs,
		CreatedAt:      utcNowText(),
	}
	data, err := json.Marshal(rec)
	if err != nil {
		return fmt.Errorf("encoding activity entry: %w", err)
	}
	return s.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("activity_log"))
		if b == nil {
			return fmt.Errorf("activity_log bucket missing")
		}
		if err := b.Put([]byte(rec.ID), data); err != nil {
			return fmt.Errorf("writing activity entry: %w", err)
		}
		return nil
	})
}

func (s *BoltStore) ListActivity(limit, offset int, clientIP, domain, action string) ([]ActivityEntryView, int, error) {
	items := make([]ActivityEntryView, 0)
	total := 0
	if err := s.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("activity_log"))
		if b == nil {
			return fmt.Errorf("activity_log bucket missing")
		}
		return b.ForEach(func(_, v []byte) error {
			var rec ActivityEntryView
			if err := json.Unmarshal(v, &rec); err != nil {
				return nil
			}
			if clientIP != "" && rec.ClientIP != clientIP {
				return nil
			}
			if domain != "" && !strings.Contains(strings.ToLower(rec.Domain), strings.ToLower(domain)) {
				return nil
			}
			if action != "" && rec.Action != action {
				return nil
			}
			total++
			items = append(items, rec)
			return nil
		})
	}); err != nil {
		return nil, 0, err
	}

	sort.Slice(items, func(i, j int) bool {
		return items[i].CreatedAt > items[j].CreatedAt
	})
	return paginateActivityViews(items, limit, offset), total, nil
}

func (s *BoltStore) ActivityCounts() (total, blocked, allowed int, err error) {
	err = s.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("activity_log"))
		if b == nil {
			return fmt.Errorf("activity_log bucket missing")
		}
		return b.ForEach(func(_, v []byte) error {
			var rec ActivityEntryView
			if err := json.Unmarshal(v, &rec); err != nil {
				return nil
			}
			total++
			switch rec.Action {
			case "blocked":
				blocked++
			case "allowed":
				allowed++
			}
			return nil
		})
	})
	return total, blocked, allowed, err
}

func (s *BoltStore) TopBlockedDomains(limit int) ([]TopBlockedDomainView, error) {
	counts := make(map[string]int)
	if err := s.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("activity_log"))
		if b == nil {
			return fmt.Errorf("activity_log bucket missing")
		}
		return b.ForEach(func(_, v []byte) error {
			var rec ActivityEntryView
			if err := json.Unmarshal(v, &rec); err != nil {
				return nil
			}
			if rec.Action == "blocked" {
				counts[rec.Domain]++
			}
			return nil
		})
	}); err != nil {
		return nil, err
	}

	items := make([]TopBlockedDomainView, 0, len(counts))
	for domain, count := range counts {
		items = append(items, TopBlockedDomainView{Domain: domain, Count: count})
	}
	sort.Slice(items, func(i, j int) bool {
		if items[i].Count == items[j].Count {
			return items[i].Domain < items[j].Domain
		}
		return items[i].Count > items[j].Count
	})
	if limit > 0 && len(items) > limit {
		items = items[:limit]
	}
	return items, nil
}

func paginateActivityViews(items []ActivityEntryView, limit, offset int) []ActivityEntryView {
	if offset >= len(items) {
		return []ActivityEntryView{}
	}
	end := offset + limit
	if end > len(items) {
		end = len(items)
	}
	return items[offset:end]
}

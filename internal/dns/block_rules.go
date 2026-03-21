package dns

import (
	"database/sql"
	"log/slog"
	"regexp"
	"strings"
	"sync"
)

type blockRuleMatcher struct {
	db *sql.DB

	once sync.Once
	err  error

	exactStmt *sql.Stmt

	mu        sync.RWMutex
	wildcards []string
	regexes   []*regexp.Regexp
}

func newBlockRuleMatcher(db *sql.DB) *blockRuleMatcher {
	return &blockRuleMatcher{db: db}
}

func (m *blockRuleMatcher) prime() error {
	m.once.Do(func() {
		if m.db == nil {
			m.err = sql.ErrConnDone
			return
		}

		stmt, err := m.db.Prepare(`SELECT EXISTS(SELECT 1 FROM block_rules WHERE enabled=1 AND rule_type='exact' AND LOWER(pattern)=LOWER($1))`)
		if err != nil {
			m.err = err
			return
		}
		m.exactStmt = stmt

		wildcards, err := m.loadPatterns("wildcard")
		if err != nil {
			m.err = err
			return
		}

		regexPatterns, err := m.loadPatterns("regex")
		if err != nil {
			m.err = err
			return
		}

		compiled := make([]*regexp.Regexp, 0, len(regexPatterns))
		for _, pattern := range regexPatterns {
			re, err := regexp.Compile("(?i)" + pattern)
			if err != nil {
				slog.Warn("ignoring invalid block regex rule", "pattern", pattern, "err", err)
				continue
			}
			compiled = append(compiled, re)
		}

		m.mu.Lock()
		m.wildcards = wildcards
		m.regexes = compiled
		m.mu.Unlock()
	})

	return m.err
}

func (m *blockRuleMatcher) loadPatterns(ruleType string) ([]string, error) {
	rows, err := m.db.Query(`SELECT pattern FROM block_rules WHERE enabled=1 AND rule_type=$1`, ruleType)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	patterns := make([]string, 0)
	for rows.Next() {
		var pattern string
		if err := rows.Scan(&pattern); err != nil {
			return nil, err
		}
		patterns = append(patterns, strings.ToLower(strings.TrimSpace(pattern)))
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return patterns, nil
}

func (m *blockRuleMatcher) matches(domain string) bool {
	if err := m.prime(); err != nil {
		slog.Error("failed to prime block rule matcher", "err", err)
		return false
	}

	domain = normaliseDomain(domain)
	if domain == "" {
		return false
	}

	var exactMatch bool
	if err := m.exactStmt.QueryRow(domain).Scan(&exactMatch); err != nil {
		slog.Warn("exact block rule lookup failed", "domain", domain, "err", err)
	} else if exactMatch {
		return true
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, pattern := range m.wildcards {
		if wildcardMatches(pattern, domain) {
			return true
		}
	}

	for _, re := range m.regexes {
		if re.MatchString(domain) {
			return true
		}
	}

	return false
}

func wildcardMatches(pattern, domain string) bool {
	pattern = strings.ToLower(strings.TrimSpace(pattern))
	domain = normaliseDomain(domain)

	if !strings.HasPrefix(pattern, "*.") {
		return false
	}

	suffix := strings.TrimPrefix(pattern, "*")
	return strings.HasSuffix(domain, suffix)
}

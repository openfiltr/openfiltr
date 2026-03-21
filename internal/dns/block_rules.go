package dns

import (
	"database/sql"
	"log/slog"
	"regexp"
	"strings"
	"sync"

	"github.com/openfiltr/openfiltr/internal/storage"
)

// blockRuleState holds the loaded state for rule matching.
type blockRuleState struct {
	exactStmt *sql.Stmt
	wildcards []string
	regexes   []*regexp.Regexp
}

// Rule types stored in the block_rules table.
const (
	ruleTypeExact    = "exact"
	ruleTypeWildcard = "wildcard"
	ruleTypeRegex    = "regex"
)

type blockRuleMatcher struct {
	db *sql.DB

	loadMu sync.Mutex   // serialises prime and Reload calls
	mu     sync.RWMutex // protects state
	state  *blockRuleState
}

func newBlockRuleMatcher(db *sql.DB) *blockRuleMatcher {
	return &blockRuleMatcher{db: db}
}

// prime loads block rules once. Subsequent calls are no-ops unless Reload has been called.
func (m *blockRuleMatcher) prime() error {
	m.loadMu.Lock()
	defer m.loadMu.Unlock()

	m.mu.RLock()
	loaded := m.state != nil
	m.mu.RUnlock()
	if loaded {
		return nil
	}

	return m.load()
}

// Reload re-reads all block rules from the database, replacing the current state.
// It can be called after rules are updated to pick up the latest patterns without
// restarting the DNS server.
func (m *blockRuleMatcher) Reload() error {
	m.loadMu.Lock()
	defer m.loadMu.Unlock()
	return m.load()
}

// Close releases the prepared exact-match statement. Call from Server.Stop().
func (m *blockRuleMatcher) Close() {
	m.loadMu.Lock()
	defer m.loadMu.Unlock()

	m.mu.Lock()
	old := m.state
	m.state = nil
	m.mu.Unlock()

	if old != nil && old.exactStmt != nil {
		_ = old.exactStmt.Close()
	}
}

// load builds a fresh blockRuleState and atomically installs it. The caller must hold loadMu.
func (m *blockRuleMatcher) load() error {
	if m.db == nil {
		return sql.ErrConnDone
	}

	stmt, err := m.db.Prepare(storage.Rebind(
		`SELECT EXISTS(SELECT 1 FROM block_rules WHERE enabled=1 AND rule_type='exact' AND LOWER(pattern)=LOWER(?))`))
	if err != nil {
		return err
	}

	wildcards, err := m.loadPatterns(ruleTypeWildcard)
	if err != nil {
		_ = stmt.Close()
		return err
	}

	regexPatterns, err := m.loadPatterns(ruleTypeRegex)
	if err != nil {
		_ = stmt.Close()
		return err
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

	newState := &blockRuleState{
		exactStmt: stmt,
		wildcards: wildcards,
		regexes:   compiled,
	}

	m.mu.Lock()
	old := m.state
	m.state = newState
	m.mu.Unlock()

	// Close the old prepared statement after swapping it out.
	if old != nil && old.exactStmt != nil {
		_ = old.exactStmt.Close()
	}

	return nil
}

func (m *blockRuleMatcher) loadPatterns(ruleType string) ([]string, error) {
	rows, err := m.db.Query(storage.Rebind(`SELECT pattern FROM block_rules WHERE enabled=1 AND rule_type=?`), ruleType)
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
		pattern = strings.TrimSpace(pattern)
		if ruleType != ruleTypeRegex {
			pattern = strings.ToLower(pattern)
		}
		patterns = append(patterns, pattern)
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

	// Take a snapshot of the state under read lock so Reload can proceed concurrently.
	m.mu.RLock()
	s := m.state
	m.mu.RUnlock()

	if s == nil {
		return false
	}

	var exactMatch bool
	if err := s.exactStmt.QueryRow(domain).Scan(&exactMatch); err != nil {
		slog.Warn("exact block rule lookup failed", "domain", domain, "err", err)
	} else if exactMatch {
		return true
	}

	for _, pattern := range s.wildcards {
		if wildcardMatches(pattern, domain) {
			return true
		}
	}

	for _, re := range s.regexes {
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

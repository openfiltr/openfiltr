package storage

import "time"

type RuleView struct {
	ID        string  `json:"id"`
	Pattern   string  `json:"pattern"`
	RuleType  string  `json:"rule_type"`
	Comment   *string `json:"comment"`
	Enabled   int     `json:"enabled"`
	CreatedBy *string `json:"created_by"`
	CreatedAt string  `json:"created_at"`
	UpdatedAt string  `json:"updated_at"`
}

type DNSEntryView struct {
	ID        string  `json:"id"`
	Host      string  `json:"host"`
	EntryType string  `json:"entry_type"`
	Value     string  `json:"value"`
	TTL       int     `json:"ttl"`
	Comment   *string `json:"comment"`
	Enabled   int     `json:"enabled"`
	CreatedBy *string `json:"created_by"`
	CreatedAt string  `json:"created_at"`
	UpdatedAt string  `json:"updated_at"`
}

type ActivityEntryView struct {
	ID             string  `json:"id"`
	ClientIP       string  `json:"client_ip"`
	Domain         string  `json:"domain"`
	QueryType      string  `json:"query_type"`
	Action         string  `json:"action"`
	RuleID         *string `json:"rule_id"`
	RuleSource     *string `json:"rule_source"`
	ResponseTimeMs *int    `json:"response_time_ms"`
	CreatedAt      string  `json:"created_at"`
}

type TopBlockedDomainView struct {
	Domain string `json:"domain"`
	Count  int    `json:"count"`
}

type UpstreamServerView struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Address   string `json:"address"`
	Protocol  string `json:"protocol"`
	Enabled   int    `json:"enabled"`
	Priority  int    `json:"priority"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type ClientView struct {
	ID             string  `json:"id"`
	Name           string  `json:"name"`
	Identifier     string  `json:"identifier"`
	IdentifierType string  `json:"identifier_type"`
	GroupID        *string `json:"group_id"`
	Comment        *string `json:"comment"`
	CreatedAt      string  `json:"created_at"`
	UpdatedAt      string  `json:"updated_at"`
}

type GroupView struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description *string `json:"description"`
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   string  `json:"updated_at"`
}

type RuleSourceView struct {
	ID            string  `json:"id"`
	Name          string  `json:"name"`
	URL           string  `json:"url"`
	Format        string  `json:"format"`
	Enabled       int     `json:"enabled"`
	LastUpdatedAt *string `json:"last_updated_at"`
	RuleCount     int     `json:"rule_count"`
	CreatedAt     string  `json:"created_at"`
	UpdatedAt     string  `json:"updated_at"`
}

type AuditEventView struct {
	ID           string  `json:"id"`
	UserID       *string `json:"user_id"`
	Action       string  `json:"action"`
	ResourceType string  `json:"resource_type"`
	ResourceID   *string `json:"resource_id"`
	Details      *string `json:"details"`
	IPAddress    *string `json:"ip_address"`
	CreatedAt    string  `json:"created_at"`
}

func utcNowText() string {
	return time.Now().UTC().Format(time.RFC3339)
}

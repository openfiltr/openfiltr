package dns

import (
	"fmt"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
)

func TestWildcardPatterns(t *testing.T) {
	tests := []struct {
		name   string
		domain string
		want   []string
	}{
		{
			name:   "subdomain yields wildcard suffixes",
			domain: "api.eu.example.com",
			want:   []string{"*.eu.example.com", "*.example.com", "*.com"},
		},
		{
			name:   "apex domain only yields parent suffix",
			domain: "example.com",
			want:   []string{"*.com"},
		},
		{
			name:   "single label has no wildcard patterns",
			domain: "localhost",
			want:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := wildcardPatterns(tt.domain)
			if fmt.Sprint(got) != fmt.Sprint(tt.want) {
				t.Fatalf("wildcardPatterns() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsBlockedMatchesExactRulesCaseInsensitively(t *testing.T) {
	srv, mock, cleanup := newMockServer(t)
	defer cleanup()

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM block_rules WHERE enabled=1 AND rule_type='exact' AND lower\(pattern\)=\$1`).
		WithArgs("example.com").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	if !srv.isBlocked("Example.COM.") {
		t.Fatal("isBlocked() = false, want true")
	}

	assertExpectations(t, mock)
}

func TestIsBlockedMatchesWildcardRulesForSubdomainsOnly(t *testing.T) {
	srv, mock, cleanup := newMockServer(t)
	defer cleanup()

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM block_rules WHERE enabled=1 AND rule_type='exact' AND lower\(pattern\)=\$1`).
		WithArgs("foo.example.com").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM block_rules WHERE enabled=1 AND rule_type='wildcard' AND lower\(pattern\) = ANY\(\$1\)`).
		WithArgs(sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	if !srv.isBlocked("foo.example.com") {
		t.Fatal("isBlocked() = false, want true")
	}

	assertExpectations(t, mock)
}

func TestIsBlockedDoesNotMatchWildcardAgainstApexDomain(t *testing.T) {
	srv, mock, cleanup := newMockServer(t)
	defer cleanup()

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM block_rules WHERE enabled=1 AND rule_type='exact' AND lower\(pattern\)=\$1`).
		WithArgs("example.com").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM block_rules WHERE enabled=1 AND rule_type='wildcard' AND lower\(pattern\) = ANY\(\$1\)`).
		WithArgs(sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	mock.ExpectQuery(`SELECT pattern FROM block_rules WHERE enabled=1 AND rule_type='regex'`).
		WillReturnRows(sqlmock.NewRows([]string{"pattern"}))

	if srv.isBlocked("example.com") {
		t.Fatal("isBlocked() = true, want false")
	}

	assertExpectations(t, mock)
}

func TestIsBlockedMatchesRegexRulesCaseInsensitively(t *testing.T) {
	srv, mock, cleanup := newMockServer(t)
	defer cleanup()

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM block_rules WHERE enabled=1 AND rule_type='exact' AND lower\(pattern\)=\$1`).
		WithArgs("api.example.com").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM block_rules WHERE enabled=1 AND rule_type='wildcard' AND lower\(pattern\) = ANY\(\$1\)`).
		WithArgs(sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	mock.ExpectQuery(`SELECT pattern FROM block_rules WHERE enabled=1 AND rule_type='regex'`).
		WillReturnRows(sqlmock.NewRows([]string{"pattern"}).AddRow(`^api\.example\.com$`))

	if !srv.isBlocked("API.Example.Com") {
		t.Fatal("isBlocked() = false, want true")
	}

	assertExpectations(t, mock)
}

func newMockServer(t *testing.T) (*Server, sqlmock.Sqlmock, func()) {
	t.Helper()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New() error = %v", err)
	}

	srv := NewServer(nil, db)
	cleanup := func() {
		_ = db.Close()
	}

	return srv, mock, cleanup
}

func assertExpectations(t *testing.T, mock sqlmock.Sqlmock) {
	t.Helper()
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations not met: %v", err)
	}
}

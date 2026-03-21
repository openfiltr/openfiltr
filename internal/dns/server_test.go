package dns

import (
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
)

func TestNormaliseDomain(t *testing.T) {
	if got := normaliseDomain(" Example.COM. "); got != "example.com" {
		t.Fatalf("normaliseDomain() = %q, want %q", got, "example.com")
	}
}

func TestServerIsBlockedMatchesExactRulesCaseInsensitively(t *testing.T) {
	srv, mock, cleanup := newMockServer(t)
	defer cleanup()

	mock.ExpectPrepare(`SELECT EXISTS\(SELECT 1 FROM block_rules WHERE enabled=1 AND rule_type='exact' AND LOWER\(pattern\)=LOWER\(\$1\)\)`)
	mock.ExpectQuery(`SELECT pattern FROM block_rules WHERE enabled=1 AND rule_type=\$1`).
		WithArgs("wildcard").
		WillReturnRows(sqlmock.NewRows([]string{"pattern"}))
	mock.ExpectQuery(`SELECT pattern FROM block_rules WHERE enabled=1 AND rule_type=\$1`).
		WithArgs("regex").
		WillReturnRows(sqlmock.NewRows([]string{"pattern"}))
	mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM block_rules WHERE enabled=1 AND rule_type='exact' AND LOWER\(pattern\)=LOWER\(\$1\)\)`).
		WithArgs("example.com").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	if !srv.isBlocked("Example.COM.") {
		t.Fatal("isBlocked() = false, want true")
	}

	assertExpectations(t, mock)
}

func TestServerIsBlockedMatchesWildcardRulesForSubdomainsOnly(t *testing.T) {
	srv, mock, cleanup := newMockServer(t)
	defer cleanup()

	mock.ExpectPrepare(`SELECT EXISTS\(SELECT 1 FROM block_rules WHERE enabled=1 AND rule_type='exact' AND LOWER\(pattern\)=LOWER\(\$1\)\)`)
	mock.ExpectQuery(`SELECT pattern FROM block_rules WHERE enabled=1 AND rule_type=\$1`).
		WithArgs("wildcard").
		WillReturnRows(sqlmock.NewRows([]string{"pattern"}).AddRow("*.example.com"))
	mock.ExpectQuery(`SELECT pattern FROM block_rules WHERE enabled=1 AND rule_type=\$1`).
		WithArgs("regex").
		WillReturnRows(sqlmock.NewRows([]string{"pattern"}))
	mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM block_rules WHERE enabled=1 AND rule_type='exact' AND LOWER\(pattern\)=LOWER\(\$1\)\)`).
		WithArgs("foo.example.com").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))
	mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM block_rules WHERE enabled=1 AND rule_type='exact' AND LOWER\(pattern\)=LOWER\(\$1\)\)`).
		WithArgs("example.com").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

	if !srv.isBlocked("foo.example.com") {
		t.Fatal("isBlocked() = false, want true")
	}
	if srv.isBlocked("example.com") {
		t.Fatal("isBlocked() = true, want false")
	}

	assertExpectations(t, mock)
}

func TestServerIsBlockedMatchesRegexRulesCaseInsensitively(t *testing.T) {
	srv, mock, cleanup := newMockServer(t)
	defer cleanup()

	mock.ExpectPrepare(`SELECT EXISTS\(SELECT 1 FROM block_rules WHERE enabled=1 AND rule_type='exact' AND LOWER\(pattern\)=LOWER\(\$1\)\)`)
	mock.ExpectQuery(`SELECT pattern FROM block_rules WHERE enabled=1 AND rule_type=\$1`).
		WithArgs("wildcard").
		WillReturnRows(sqlmock.NewRows([]string{"pattern"}))
	mock.ExpectQuery(`SELECT pattern FROM block_rules WHERE enabled=1 AND rule_type=\$1`).
		WithArgs("regex").
		WillReturnRows(sqlmock.NewRows([]string{"pattern"}).AddRow(`^api\.example\.com$`))
	mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM block_rules WHERE enabled=1 AND rule_type='exact' AND LOWER\(pattern\)=LOWER\(\$1\)\)`).
		WithArgs("api.example.com").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

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

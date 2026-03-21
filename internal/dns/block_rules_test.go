package dns

import (
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
)

func TestBlockRuleMatcherMatchesExactCaseInsensitive(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New() error = %v", err)
	}
	defer db.Close()

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

	matcher := newBlockRuleMatcher(db)
	if !matcher.matches("Example.COM") {
		t.Fatal("matches() = false, want true for exact rule")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations not met: %v", err)
	}
}

func TestBlockRuleMatcherMatchesWildcardCaseInsensitive(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New() error = %v", err)
	}
	defer db.Close()

	mock.ExpectPrepare(`SELECT EXISTS\(SELECT 1 FROM block_rules WHERE enabled=1 AND rule_type='exact' AND LOWER\(pattern\)=LOWER\(\$1\)\)`)
	mock.ExpectQuery(`SELECT pattern FROM block_rules WHERE enabled=1 AND rule_type=\$1`).
		WithArgs("wildcard").
		WillReturnRows(sqlmock.NewRows([]string{"pattern"}).AddRow("*.example.com"))
	mock.ExpectQuery(`SELECT pattern FROM block_rules WHERE enabled=1 AND rule_type=\$1`).
		WithArgs("regex").
		WillReturnRows(sqlmock.NewRows([]string{"pattern"}))
	mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM block_rules WHERE enabled=1 AND rule_type='exact' AND LOWER\(pattern\)=LOWER\(\$1\)\)`).
		WithArgs("sub.example.com").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))
	mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM block_rules WHERE enabled=1 AND rule_type='exact' AND LOWER\(pattern\)=LOWER\(\$1\)\)`).
		WithArgs("example.com").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

	matcher := newBlockRuleMatcher(db)
	if !matcher.matches("Sub.Example.Com") {
		t.Fatal("matches() = false, want true for wildcard rule")
	}
	if matcher.matches("example.com") {
		t.Fatal("matches() = true, want false for bare domain without subdomain")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations not met: %v", err)
	}
}

func TestBlockRuleMatcherMatchesRegexCaseInsensitive(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New() error = %v", err)
	}
	defer db.Close()

	mock.ExpectPrepare(`SELECT EXISTS\(SELECT 1 FROM block_rules WHERE enabled=1 AND rule_type='exact' AND LOWER\(pattern\)=LOWER\(\$1\)\)`)
	mock.ExpectQuery(`SELECT pattern FROM block_rules WHERE enabled=1 AND rule_type=\$1`).
		WithArgs("wildcard").
		WillReturnRows(sqlmock.NewRows([]string{"pattern"}))
	mock.ExpectQuery(`SELECT pattern FROM block_rules WHERE enabled=1 AND rule_type=\$1`).
		WithArgs("regex").
		WillReturnRows(sqlmock.NewRows([]string{"pattern"}).AddRow(`^ads[0-9]+\.example\.com$`))
	mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM block_rules WHERE enabled=1 AND rule_type='exact' AND LOWER\(pattern\)=LOWER\(\$1\)\)`).
		WithArgs("ads42.example.com").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

	matcher := newBlockRuleMatcher(db)
	if !matcher.matches("Ads42.Example.Com") {
		t.Fatal("matches() = false, want true for regex rule")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations not met: %v", err)
	}
}

func TestWildcardMatchesRequiresSubdomain(t *testing.T) {
	if !wildcardMatches("*.example.com", "foo.example.com") {
		t.Fatal("wildcardMatches() = false, want true for subdomain")
	}
	if !wildcardMatches("*.example.com", "a.b.example.com") {
		t.Fatal("wildcardMatches() = false, want true for nested subdomain")
	}
	if wildcardMatches("*.example.com", "example.com") {
		t.Fatal("wildcardMatches() = true, want false for bare domain")
	}
}

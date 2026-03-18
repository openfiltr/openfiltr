package api

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestExportPayloadIncludesSchemaVersion(t *testing.T) {
	data, err := yaml.Marshal(exportPayload{Version: configExportVersion})
	if err != nil {
		t.Fatalf("yaml.Marshal() error = %v", err)
	}
	want := fmt.Sprintf("version: %d\n", configExportVersion)
	if !strings.HasPrefix(string(data), want) {
		t.Fatalf("export YAML = %q, want prefix %q", string(data), want)
	}
}

func TestImportConfigRejectsUnsupportedVersion(t *testing.T) {
	h := &Handler{}
	req := httptest.NewRequest(http.MethodPost, "/api/v1/config/import", strings.NewReader("version: 2\n"))
	w := httptest.NewRecorder()

	h.ImportConfig(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("ImportConfig() status = %d, want %d", w.Code, http.StatusBadRequest)
	}
	if body := w.Body.String(); !strings.Contains(body, "unsupported config version 2") {
		t.Fatalf("ImportConfig() body = %q, want unsupported version error", body)
	}
}

func TestImportConfigRejectsMissingVersion(t *testing.T) {
	h := &Handler{}
	req := httptest.NewRequest(http.MethodPost, "/api/v1/config/import", strings.NewReader("block_rules: []\n"))
	w := httptest.NewRecorder()

	h.ImportConfig(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("ImportConfig() status = %d, want %d", w.Code, http.StatusBadRequest)
	}
	if body := w.Body.String(); !strings.Contains(body, "missing required top-level version") {
		t.Fatalf("ImportConfig() body = %q, want missing version error", body)
	}
}

func TestNormaliseRowValueConvertsBytesToString(t *testing.T) {
	got := normaliseRowValue([]byte("example"))
	if got != "example" {
		t.Fatalf("normaliseRowValue() = %#v, want %q", got, "example")
	}
}

func TestStringValueAcceptsBytes(t *testing.T) {
	row := map[string]interface{}{"pattern": []byte("ads.example")}
	if got := stringValue(row, "pattern"); got != "ads.example" {
		t.Fatalf("stringValue() = %q, want %q", got, "ads.example")
	}
}

func TestNullableStringValuePreservesNil(t *testing.T) {
	row := map[string]interface{}{"comment": nil}
	if got := nullableStringValue(row, "comment"); got != nil {
		t.Fatalf("nullableStringValue() = %#v, want nil", got)
	}
}

func TestIntValueAcceptsStringAndFloat(t *testing.T) {
	row := map[string]interface{}{
		"enabled":  "1",
		"priority": float64(7),
	}
	if got := intValue(row, "enabled", 0); got != 1 {
		t.Fatalf("intValue(enabled) = %d, want 1", got)
	}
	if got := intValue(row, "priority", 0); got != 7 {
		t.Fatalf("intValue(priority) = %d, want 7", got)
	}
}

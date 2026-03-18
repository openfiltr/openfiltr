package api

import "testing"

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

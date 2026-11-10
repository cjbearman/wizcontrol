package homeassistant

import "testing"

func TestJsonEscapeStr(t *testing.T) {
	str := "test string"
	escaped := jsonEscapeStr(str)
	if escaped != `"test string"` {
		t.Errorf("Expected %q, got %q", `"test string"`, escaped)
	}
}

func TestJsonEscapeStrNoQuotes(t *testing.T) {
	str := "test string"
	escaped := jsonEscapeStrNoQuotes(str)
	if escaped != `test string` {
		t.Errorf("Expected %q, got %q", `test string`, escaped)
	}
}

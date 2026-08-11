package tools

import (
	"regexp"
	"testing"
)

var uuidRegex = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)

func TestGenerateUUID(t *testing.T) {
	id1 := GenerateUUID()
	id2 := GenerateUUID()

	if !uuidRegex.MatchString(id1) {
		t.Errorf("GenerateUUID() output %q does not match RFC4122 v4 UUID format", id1)
	}

	if !uuidRegex.MatchString(id2) {
		t.Errorf("GenerateUUID() output %q does not match RFC4122 v4 UUID format", id2)
	}

	if id1 == id2 {
		t.Errorf("expected generated UUIDs to be unique, got duplicate %q", id1)
	}
}

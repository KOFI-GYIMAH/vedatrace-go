package vedatrace

import "testing"

func TestRedact_noFields(t *testing.T) {
	meta := LogMetadata{"password": "secret"}
	out := redact(meta, nil)
	if out["password"] != "secret" {
		t.Error("should not redact when no fields configured")
	}
}

func TestRedact_shallowField(t *testing.T) {
	meta := LogMetadata{"password": "secret", "user": "alice"}
	out := redact(meta, []string{"password"})
	if out["password"] != "[REDACTED]" {
		t.Errorf("expected [REDACTED], got %v", out["password"])
	}
	if out["user"] != "alice" {
		t.Error("non-redacted field should be unchanged")
	}
}

func TestRedact_nestedField(t *testing.T) {
	meta := LogMetadata{
		"card": map[string]any{
			"number": "4111111111111111",
			"cvv":    "123",
		},
	}
	out := redact(meta, []string{"card.cvv"})
	card, ok := out["card"].(LogMetadata)
	if !ok {
		// may be map[string]any
		raw, _ := out["card"].(map[string]any)
		if raw["cvv"] != "[REDACTED]" {
			t.Errorf("nested cvv should be redacted, got %v", raw["cvv"])
		}
		if raw["number"] != "4111111111111111" {
			t.Error("card.number should not be redacted")
		}
		return
	}
	if card["cvv"] != "[REDACTED]" {
		t.Errorf("nested cvv should be redacted, got %v", card["cvv"])
	}
}

func TestRedact_doesNotMutateOriginal(t *testing.T) {
	meta := LogMetadata{"secret": "value"}
	_ = redact(meta, []string{"secret"})
	if meta["secret"] != "value" {
		t.Error("original metadata was mutated")
	}
}

func TestRedact_missingField(t *testing.T) {
	meta := LogMetadata{"user": "alice"}
	out := redact(meta, []string{"password"})
	if _, ok := out["password"]; ok {
		t.Error("should not add missing field")
	}
}

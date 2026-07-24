package anima

import (
	"encoding/json"
	"strings"
	"testing"
)

// TestVoiceUnmarshalCatalog locks the vendor-neutral catalog wire shape: the
// camelCase JSON tags (useCases, sampleUrl) must map, descriptors/useCases must
// decode, and a non-English language must round-trip (multilingual catalog).
func TestVoiceUnmarshalCatalog(t *testing.T) {
	raw := `{
		"voices": [
			{
				"id": "celeste",
				"name": "Celeste",
				"gender": "female",
				"accent": "Castilian",
				"age": "adult",
				"descriptors": ["warm", "smooth"],
				"useCases": ["support", "sales"],
				"language": "es",
				"sampleUrl": "https://api.useanima.sh/v1/voice/catalog/celeste/sample"
			}
		]
	}`

	var list VoiceList
	if err := json.Unmarshal([]byte(raw), &list); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(list.Voices) != 1 {
		t.Fatalf("want 1 voice, got %d", len(list.Voices))
	}
	v := list.Voices[0]
	if v.ID != "celeste" || v.Name != "Celeste" {
		t.Errorf("id/name mismatch: %q %q", v.ID, v.Name)
	}
	if v.Gender != "female" {
		t.Errorf("gender = %q, want female", v.Gender)
	}
	if v.Language != "es" {
		t.Errorf("language = %q, want es (multilingual catalog)", v.Language)
	}
	if len(v.Descriptors) != 2 || v.Descriptors[0] != "warm" {
		t.Errorf("descriptors = %v", v.Descriptors)
	}
	if len(v.UseCases) != 2 || v.UseCases[1] != "sales" {
		t.Errorf("useCases = %v", v.UseCases)
	}
	if v.SampleURL != "https://api.useanima.sh/v1/voice/catalog/celeste/sample" {
		t.Errorf("sampleUrl = %q", v.SampleURL)
	}
}

// TestVoiceMarshalTags locks the JSON tags that an unmarshal test cannot:
// omitempty on the optionals, and the EXACT (case-sensitive) key spelling.
// Go's json unmarshal is case-insensitive, so a sampleUrl->sampleURL tag
// regression slips past a decode test — but marshal is case-exact.
func TestVoiceMarshalTags(t *testing.T) {
	// Only required fields set; optionals left empty.
	minimal, err := json.Marshal(Voice{
		ID:          "thalia",
		Name:        "Thalia",
		Gender:      "neutral",
		Descriptors: []string{},
		UseCases:    []string{},
		Language:    "en",
	})
	if err != nil {
		t.Fatalf("marshal minimal: %v", err)
	}
	got := string(minimal)
	// omitempty: empty optionals must be absent from the output.
	for _, key := range []string{`"accent"`, `"age"`, `"sampleUrl"`} {
		if strings.Contains(got, key) {
			t.Errorf("empty optional %s should be omitted, got: %s", key, got)
		}
	}
	// Required keys must be present with exact case.
	for _, key := range []string{`"id"`, `"name"`, `"gender"`, `"descriptors"`, `"useCases"`, `"language"`} {
		if !strings.Contains(got, key) {
			t.Errorf("required key %s missing, got: %s", key, got)
		}
	}

	// When set, the optionals marshal to their exact camelCase keys — not
	// sampleURL / sample_url / usecases (regressions a decode test can't catch).
	full, err := json.Marshal(Voice{
		ID:          "celeste",
		Name:        "Celeste",
		Gender:      "female",
		Accent:      "Castilian",
		Descriptors: []string{"warm"},
		UseCases:    []string{"support"},
		Language:    "es",
		SampleURL:   "https://api.useanima.sh/v1/voice/catalog/celeste/sample",
	})
	if err != nil {
		t.Fatalf("marshal full: %v", err)
	}
	got2 := string(full)
	if !strings.Contains(got2, `"sampleUrl"`) || !strings.Contains(got2, `"accent"`) {
		t.Errorf(`expected exact-case "sampleUrl" and "accent" when set, got: %s`, got2)
	}
	for _, bad := range []string{`"sampleURL"`, `"sample_url"`, `"usecases"`, `"use_cases"`} {
		if strings.Contains(got2, bad) {
			t.Errorf("mis-cased/shaped tag %s regressed, got: %s", bad, got2)
		}
	}
}

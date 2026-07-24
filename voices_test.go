package anima

import (
	"encoding/json"
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

// TestVoiceMinimalOmitsOptionalFields verifies a voice without optional fields
// (accent/age/sampleUrl) decodes cleanly to zero values.
func TestVoiceMinimalOmitsOptionalFields(t *testing.T) {
	raw := `{"id":"thalia","name":"Thalia","gender":"neutral","descriptors":[],"useCases":[],"language":"en"}`
	var v Voice
	if err := json.Unmarshal([]byte(raw), &v); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if v.Accent != "" || v.Age != "" || v.SampleURL != "" {
		t.Errorf("expected empty optionals, got accent=%q age=%q sampleUrl=%q", v.Accent, v.Age, v.SampleURL)
	}
}

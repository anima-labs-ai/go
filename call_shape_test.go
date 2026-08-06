package anima

import (
	"reflect"
	"sort"
	"strings"
	"testing"
)

// The call structs are a wire contract with the API's voice schemas.
//
// Go hides this class of drift better than the other SDKs, which is why it
// lasted longest here. `Call.Tier` was a plain string with `json:"tier"`, the
// API has never sent a tier, and encoding/json leaves a missing field at its
// zero value without complaining. So `call.Tier` was "" on every call, forever,
// and cmd/test-voice printed "Tier: " on every line. Python raised on the same
// payload; TypeScript typed it as present and handed back undefined.
//
// The field lists below are spelled out from packages/contracts/src/schemas/
// voice.ts rather than derived from the structs — a check that reads the same
// type it is checking cannot fail when that type is wrong.

var liveCallFields = []string{
	"id", "agentId", "phoneIdentityId", "direction", "state", "from", "to",
	"startedAt", "answeredAt", "endedAt", "endReason", "durationSeconds", "createdAt",
}

var liveCreateCallInputFields = []string{"to", "agentId", "greeting", "fromNumber"}

var liveCreateCallOutputFields = []string{"callId", "state", "from", "to", "direction"}

// jsonFields returns the wire names a struct serializes, tags stripped of
// options like ",omitempty".
func jsonFields(t reflect.Type) []string {
	names := make([]string, 0, t.NumField())
	for i := 0; i < t.NumField(); i++ {
		tag := t.Field(i).Tag.Get("json")
		if tag == "" || tag == "-" {
			continue
		}
		names = append(names, strings.Split(tag, ",")[0])
	}
	sort.Strings(names)
	return names
}

func assertFields(t *testing.T, name string, got reflect.Type, want []string) {
	t.Helper()
	sorted := append([]string(nil), want...)
	sort.Strings(sorted)
	if have := jsonFields(got); !reflect.DeepEqual(have, sorted) {
		// Both directions matter: a field the API returns but the struct omits is
		// data a caller cannot reach, and one the struct declares but the API
		// never sends decodes to the zero value and reads as real.
		t.Errorf("%s fields = %v, contract = %v", name, have, sorted)
	}
}

func TestCallShapesMatchContract(t *testing.T) {
	assertFields(t, "Call", reflect.TypeOf(Call{}), liveCallFields)
	assertFields(t, "CreateCallParams", reflect.TypeOf(CreateCallParams{}), liveCreateCallInputFields)
	assertFields(t, "CreateCallOutput", reflect.TypeOf(CreateCallOutput{}), liveCreateCallOutputFields)
}

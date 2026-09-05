package anima

import (
	"reflect"
	"sort"
	"strings"
	"testing"
)

// PhoneIdentity is a wire contract with the API's phone schemas.
//
// `PhoneIdentity.Provider` was a plain string with `json:"provider"`, and the
// API stopped sending it: both PhoneIdentityOutput
// (packages/contracts/src/schemas/agent.ts) and PhoneProvisionOutput
// (.../phone.ts) dropped the field, and oRPC strips what the schema does not
// declare. As with Call.Tier, encoding/json leaves a missing field at its zero
// value without complaining, so `phone.Provider` was "" on every number,
// forever — while python raised on the same payload.
//
// It went deliberately: naming the carrier on a product surface is what the
// vendor-neutrality rule forbids. So the PhoneProvider type and its
// PhoneProviderTelnyx constant went with the field.
//
// The reverse direction cost something too: `voiceId` has been on
// PhoneIdentityOutput since the per-number voice landed and this struct never
// declared it, so callers could not reach a field the API was already sending.
//
// One struct serves both routes — PhonesService.Provision and .List both
// decode into PhoneIdentity — so the field list below is PhoneIdentityOutput's.
// A provision response simply leaves VoiceID nil.
//
// Spelled out from the contract rather than derived from the struct: a check
// that reads the same type it is checking cannot fail when that type is wrong.
var livePhoneIdentityFields = []string{
	"id", "phoneNumber", "providerId", "capabilities",
	"tenDlcStatus", "isPrimary", "voiceId", "createdAt",
}

func TestPhoneShapesMatchContract(t *testing.T) {
	assertFields(t, "PhoneIdentity", reflect.TypeOf(PhoneIdentity{}), livePhoneIdentityFields)
}

// The SMS conversation surface, added when the phone contracts gained
// listIdentities and smsThreadStats. Same rule as livePhoneIdentityFields:
// transcribed from packages/contracts/src/schemas/phone.ts, never derived from
// the structs they check.

// jsonFieldsDeep flattens embedded structs the way encoding/json does.
//
// jsonFields does not: an embedded field carries no json tag, so it is skipped
// whole. Checking PhoneIdentityListItem with it would have compared three
// fields against eleven and passed on any change to the eight it never saw —
// a guard that cannot fail.
func jsonFieldsDeep(t reflect.Type) []string {
	names := []string{}
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if f.Anonymous && f.Type.Kind() == reflect.Struct && f.Tag.Get("json") == "" {
			names = append(names, jsonFieldsDeep(f.Type)...)
			continue
		}
		tag := f.Tag.Get("json")
		if tag == "" || tag == "-" {
			continue
		}
		names = append(names, strings.Split(tag, ",")[0])
	}
	sort.Strings(names)
	return names
}

func assertFieldsDeep(t *testing.T, name string, got reflect.Type, want []string) {
	t.Helper()
	sorted := append([]string(nil), want...)
	sort.Strings(sorted)
	if have := jsonFieldsDeep(got); !reflect.DeepEqual(have, sorted) {
		t.Errorf("%s fields = %v, contract = %v", name, have, sorted)
	}
}

var livePhoneIdentityListFields = append(
	append([]string(nil), livePhoneIdentityFields...),
	"agentId", "agentName", "agentSlug",
)

var liveSMSThreadFields = []string{
	"threadId", "agentId", "participantAddress", "agentAddress",
	"lastMessageAt", "lastMessageSnippet", "lastMessageDirection",
	"messageCount", "unreadCount",
}

// {items, total, hasMore} — offset-paged, so neither the {items, pagination}
// nor the {items, nextCursor} envelope. Asserted because decoding this shape
// as a Page[T] yields HasMore false on page one and silently truncates.
var liveSMSThreadListFields = []string{"items", "total", "hasMore"}

var liveSMSThreadDetailFields = []string{
	"threadId", "agentId", "participantAddress", "agentAddress",
	"messages", "messageCount", "hasMore",
}

var liveSMSThreadStatFields = []string{
	"agentId", "conversations", "unread", "lastMessageAt",
}

var liveSMSSuppressionFields = []string{
	"id", "phoneNumber", "agentId", "reason", "source", "createdAt",
}

var liveUnsuppressSMSResultFields = []string{"phoneNumber", "removed"}

var liveInboxFields = []string{
	"id", "email", "domain", "localPart", "displayName", "agentId", "createdAt",
}

var liveInboxListItemFields = append(
	append([]string(nil), liveInboxFields...),
	"agentName", "unreadCount",
)

func TestSMSAndPhoneListShapesMatchContract(t *testing.T) {
	assertFieldsDeep(t, "PhoneIdentityListItem", reflect.TypeOf(PhoneIdentityListItem{}), livePhoneIdentityListFields)
	assertFieldsDeep(t, "SMSThread", reflect.TypeOf(SMSThread{}), liveSMSThreadFields)
	assertFieldsDeep(t, "SMSThreadList", reflect.TypeOf(SMSThreadList{}), liveSMSThreadListFields)
	assertFieldsDeep(t, "SMSThreadDetail", reflect.TypeOf(SMSThreadDetail{}), liveSMSThreadDetailFields)
	assertFieldsDeep(t, "SMSThreadStat", reflect.TypeOf(SMSThreadStat{}), liveSMSThreadStatFields)
	assertFieldsDeep(t, "SMSSuppression", reflect.TypeOf(SMSSuppression{}), liveSMSSuppressionFields)
	assertFieldsDeep(t, "UnsuppressSMSResult", reflect.TypeOf(UnsuppressSMSResult{}), liveUnsuppressSMSResultFields)
	// agentId is a plain string here, not *string: Inbox.agentId is NOT NULL.
	assertFieldsDeep(t, "Inbox", reflect.TypeOf(Inbox{}), liveInboxFields)
	assertFieldsDeep(t, "InboxListItem", reflect.TypeOf(InboxListItem{}), liveInboxListItemFields)
}

// Every 10DLC status the contract declares.
//
// UNREGISTERED has been in the contract since anima #314 (2026-07-17), where it
// is "the state every newly provisioned US long code starts in", but this
// package declared only the other four. Go does not reject an undeclared value
// — TenDLCStatus is a string alias, so the field decoded fine — but any code
// comparing against the constants never matched the one value a fresh US
// number actually carries. The python SDK, which validates, raised outright on
// the same payload.
//
// The drift canary cannot catch this class of gap: it diffs the pinned commit
// against HEAD, and this landed before the pin.
var liveTenDLCStatuses = []string{
	"PENDING", "REGISTERED", "REJECTED", "NOT_REQUIRED", "UNREGISTERED",
}

func TestTenDLCStatusConstantsCoverTheContract(t *testing.T) {
	declared := []TenDLCStatus{
		TenDLCStatusPending,
		TenDLCStatusRegistered,
		TenDLCStatusRejected,
		TenDLCStatusNotRequired,
		TenDLCStatusUnregistered,
	}
	got := make([]string, 0, len(declared))
	for _, s := range declared {
		got = append(got, string(s))
	}
	sort.Strings(got)
	want := append([]string(nil), liveTenDLCStatuses...)
	sort.Strings(want)
	if !reflect.DeepEqual(got, want) {
		t.Errorf("TenDLCStatus constants = %v, contract = %v", got, want)
	}
}

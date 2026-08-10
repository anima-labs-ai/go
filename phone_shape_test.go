package anima

import (
	"reflect"
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

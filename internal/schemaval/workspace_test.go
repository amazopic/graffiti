package schemaval

import "testing"

func TestValidateRegistry(t *testing.T) {
	ok := []byte(`{"version":"1","name":"ws","generated_at":"T","members":[{"alias":"a","path":"../a","map_hash":"h"}]}`)
	if err := ValidateRegistryBytes(ok); err != nil {
		t.Fatalf("valid registry rejected: %v", err)
	}
	for _, bad := range [][]byte{
		[]byte(`{"name":"ws"}`),                                   // missing version/members
		[]byte(`{"version":"1","name":"ws","members":"nope"}`),    // members not array
		[]byte(`{"version":"1","name":"ws","members":[{"x":1}]}`), // member missing alias
	} {
		if err := ValidateRegistryBytes(bad); err == nil {
			t.Errorf("expected error for %s", bad)
		}
	}
}

func TestValidateOverlay(t *testing.T) {
	ok := []byte(`{"version":"1","generated_at":"T","source_hashes":{"a":"h"},"links":[{"from":"a::x","to":"b::y","relation":"calls","confidence":"EXTRACTED","via":"explicit"}]}`)
	if err := ValidateOverlayBytes(ok); err != nil {
		t.Fatalf("valid overlay rejected: %v", err)
	}
	bad := []byte(`{"version":"1","links":[{"from":"a::x","relation":"calls"}]}`) // link missing to/confidence
	if err := ValidateOverlayBytes(bad); err == nil {
		t.Error("expected error for malformed link")
	}
}

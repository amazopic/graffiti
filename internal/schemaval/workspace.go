package schemaval

import (
	"encoding/json"
	"fmt"
)

// ValidateRegistryBytes checks workspace.json's required shape (structural; the
// published schema/workspace.schema.json is the informative contract).
func ValidateRegistryBytes(b []byte) error {
	var r struct {
		Version *string `json:"version"`
		Name    *string `json:"name"`
		Members *[]struct {
			Alias *string `json:"alias"`
			Path  *string `json:"path"`
		} `json:"members"`
	}
	if err := json.Unmarshal(b, &r); err != nil {
		return fmt.Errorf("registry: %w", err)
	}
	if r.Version == nil || r.Name == nil || r.Members == nil {
		return fmt.Errorf("registry: missing required version/name/members")
	}
	for i, m := range *r.Members {
		if m.Alias == nil || *m.Alias == "" || m.Path == nil || *m.Path == "" {
			return fmt.Errorf("registry: member %d missing alias/path", i)
		}
	}
	return nil
}

// ValidateOverlayBytes checks overlay.json's required shape.
func ValidateOverlayBytes(b []byte) error {
	type link struct {
		From       *string `json:"from"`
		To         *string `json:"to"`
		Relation   *string `json:"relation"`
		Confidence *string `json:"confidence"`
	}
	var o struct {
		Version *string `json:"version"`
		Links   *[]link `json:"links"`
	}
	if err := json.Unmarshal(b, &o); err != nil {
		return fmt.Errorf("overlay: %w", err)
	}
	if o.Version == nil || o.Links == nil {
		return fmt.Errorf("overlay: missing required version/links")
	}
	valid := map[string]bool{"EXTRACTED": true, "INFERRED": true, "AMBIGUOUS": true}
	for i, l := range *o.Links {
		if l.From == nil || l.To == nil || l.Relation == nil || l.Confidence == nil {
			return fmt.Errorf("overlay: link %d missing from/to/relation/confidence", i)
		}
		if !valid[*l.Confidence] {
			return fmt.Errorf("overlay: link %d bad confidence %q", i, *l.Confidence)
		}
	}
	return nil
}

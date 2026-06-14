package schema_test

import (
	"encoding/json"
	"os"
	"testing"
)

func TestMapSchemaIsValidJSON(t *testing.T) {
	b, err := os.ReadFile("map.schema.json")
	if err != nil {
		t.Fatalf("read schema: %v", err)
	}
	var doc map[string]any
	if err := json.Unmarshal(b, &doc); err != nil {
		t.Fatalf("schema is not valid JSON: %v", err)
	}
	if doc["$schema"] != "https://json-schema.org/draft/2020-12/schema" {
		t.Fatalf("$schema = %v, want draft 2020-12", doc["$schema"])
	}
	props, ok := doc["properties"].(map[string]any)
	if !ok {
		t.Fatalf("schema has no properties object")
	}
	for _, key := range []string{"version", "generated_at", "root", "nodes", "edges", "communities"} {
		if _, ok := props[key]; !ok {
			t.Fatalf("schema missing top-level property %q", key)
		}
	}
}

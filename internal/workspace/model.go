// Package workspace federates N independent per-project graffiti graphs into a
// thin computed overlay WITHOUT merging them (spec §16). Every per-project
// map.json stays unchanged and authoritative; this package only reads them and
// computes cross-project links. Alias-qualified ids ("alias::nodeid") exist only
// in the overlay and the in-memory federated index — never written back.
package workspace

// SchemaVersion is stamped into workspace.json and overlay.json.
const SchemaVersion = "1"

// WorkspaceDir is the per-root directory holding the registry and derived overlay.
const WorkspaceDir = ".graffiti-workspace"

const (
	registryFile = "workspace.json"
	overlayFile  = "overlay.json"
	linksFile    = "links"
)

// Member is one federated project: a display alias, a path relative to the
// workspace root, and the sha256 of its map.json when last seen.
type Member struct {
	Alias   string `json:"alias"`
	Path    string `json:"path"`     // relative to the workspace root, slash-separated
	MapHash string `json:"map_hash"` // sha256 hex of the member's .graffiti/map.json
}

// Registry is the committable workspace.json (no graph data — pointers + intent).
type Registry struct {
	Version     string   `json:"version"`
	Name        string   `json:"name"`
	GeneratedAt string   `json:"generated_at"` // RFC3339
	Members     []Member `json:"members"`      // sorted by alias
}

// Link is a cross-project edge in the derived overlay. Endpoints are alias::id.
// Relation/Confidence reuse the §6 vocabularies verbatim (no new enum values).
type Link struct {
	From       string `json:"from"`       // "alias::nodeid"
	To         string `json:"to"`         // "alias::nodeid"
	Relation   string `json:"relation"`   // graph.Relation value
	Confidence string `json:"confidence"` // graph.Confidence value
	Via        string `json:"via"`        // discovery provenance, e.g. "explicit"
}

// Overlay is the derived .graffiti-workspace/overlay.json (recomputable cache).
type Overlay struct {
	Version      string            `json:"version"`
	GeneratedAt  string            `json:"generated_at"`  // RFC3339
	SourceHashes map[string]string `json:"source_hashes"` // alias -> map_hash used
	Links        []Link            `json:"links"`         // confident; sorted (from,to,relation)
	Ambiguous    []Link            `json:"ambiguous"`     // surfaced for review, never traversed
}

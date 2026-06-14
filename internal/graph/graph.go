// Package graph defines graffiti's directed knowledge-graph data model (spec §6).
// It performs no I/O and no parsing.
package graph

// SchemaVersion is the version string stamped into every Document and matched by
// the published schema/map.schema.json.
const SchemaVersion = "1"

// UnclusteredCommunity is the Community value for a node before clustering (spec §6).
const UnclusteredCommunity = -1

// Kind enumerates the node kinds (spec §6).
type Kind string

const (
	KindFunction Kind = "function"
	KindMethod   Kind = "method"
	KindClass    Kind = "class"
	KindModule   Kind = "module"
	KindFile     Kind = "file"
	KindDoc      Kind = "doc"
	KindConcept  Kind = "concept"
)

// ValidKinds is the closed set of allowed kinds.
var ValidKinds = map[Kind]bool{
	KindFunction: true, KindMethod: true, KindClass: true, KindModule: true,
	KindFile: true, KindDoc: true, KindConcept: true,
}

// Relation enumerates the edge relations (spec §6).
type Relation string

const (
	RelCalls      Relation = "calls"
	RelImports    Relation = "imports"
	RelInherits   Relation = "inherits"
	RelImplements Relation = "implements"
	RelReferences Relation = "references"
	RelContains   Relation = "contains"
)

// ValidRelations is the closed set of allowed relations.
var ValidRelations = map[Relation]bool{
	RelCalls: true, RelImports: true, RelInherits: true, RelImplements: true,
	RelReferences: true, RelContains: true,
}

// Confidence enumerates the edge confidence ladder (spec §5/§6).
type Confidence string

const (
	ConfExtracted Confidence = "EXTRACTED"
	ConfInferred  Confidence = "INFERRED"
	ConfAmbiguous Confidence = "AMBIGUOUS"
)

// ValidConfidences is the closed set of allowed confidence values.
var ValidConfidences = map[Confidence]bool{
	ConfExtracted: true, ConfInferred: true, ConfAmbiguous: true,
}

// Node is a vertex in the directed graph (spec §6).
type Node struct {
	ID        string `json:"id"`
	Label     string `json:"label"`
	Kind      Kind   `json:"kind"`
	File      string `json:"file"`
	Line      int    `json:"line"`      // 1-based
	Community int    `json:"community"` // -1 before clustering
}

// Edge is a directed edge in the graph (spec §6).
type Edge struct {
	From       string     `json:"from"`
	To         string     `json:"to"`
	Relation   Relation   `json:"relation"`
	Confidence Confidence `json:"confidence"`
}

// Community is a cluster of nodes (populated by a later plan; empty in Plan 1).
type Community struct {
	ID      int      `json:"id"`
	Label   string   `json:"label"`
	Members []string `json:"members"`
}

// Document is the on-disk shape of .graffiti/map.json (spec §6).
type Document struct {
	Version     string      `json:"version"`
	GeneratedAt string      `json:"generated_at"` // RFC3339, stamped by build.Assemble
	Root        string      `json:"root"`
	Nodes       []Node      `json:"nodes"`
	Edges       []Edge      `json:"edges"`
	Communities []Community `json:"communities"`

	// nodeHighWaterMark tracks the maximum node count ever committed via Merge,
	// enabling the anti-shrink guard to detect out-of-band pruning.
	nodeHighWaterMark int
}

// NewDocument returns an empty Document with non-nil slices and the current schema version.
func NewDocument(root string) *Document {
	return &Document{
		Version:     SchemaVersion,
		Root:        root,
		Nodes:       []Node{},
		Edges:       []Edge{},
		Communities: []Community{},
	}
}

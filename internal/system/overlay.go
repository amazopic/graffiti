package system

import (
	"os"
	"path/filepath"
)

// OLink is the serialized form of a SystemLink (service-level, JSON-friendly).
type OLink struct {
	From       string `json:"from"`
	To         string `json:"to"`
	Kind       string `json:"kind"`
	Key        string `json:"key"`
	FromNode   string `json:"from_node,omitempty"`
	ToNode     string `json:"to_node,omitempty"`
	Confidence string `json:"confidence"`
	Via        string `json:"via,omitempty"`
}

// OEndpoint is the serialized form of a Dangling/Orphan entry.
type OEndpoint struct {
	Service string `json:"service"`
	Kind    string `json:"kind"`
	Key     string `json:"key"`
	Display string `json:"display,omitempty"`
}

// Overlay is the derived, recomputable cross-service overlay
// (.graffiti-system/overlay.json). Add it to .gitignore.
type Overlay struct {
	Version     string      `json:"version"`
	GeneratedAt string      `json:"generated_at"`
	Links       []OLink     `json:"links"`
	Ambiguous   []OLink     `json:"ambiguous"`
	Dangling    []OEndpoint `json:"dangling"`
	Orphans     []OEndpoint `json:"orphans"`
}

func toOLink(l SystemLink) OLink {
	return OLink{
		From: l.FromService, To: l.ToService, Kind: string(l.Kind), Key: l.Key,
		FromNode: l.FromNode, ToNode: l.ToNode, Confidence: string(l.Confidence), Via: l.Via,
	}
}

// OverlayFromMatch builds a serializable Overlay from a MatchResult.
func OverlayFromMatch(res MatchResult, generatedAt string) *Overlay {
	ov := &Overlay{Version: Version, GeneratedAt: generatedAt,
		Links: []OLink{}, Ambiguous: []OLink{}, Dangling: []OEndpoint{}, Orphans: []OEndpoint{}}
	for _, l := range res.Links {
		ov.Links = append(ov.Links, toOLink(l))
	}
	for _, l := range res.Ambiguous {
		ov.Ambiguous = append(ov.Ambiguous, toOLink(l))
	}
	for _, d := range res.Dangling {
		ov.Dangling = append(ov.Dangling, OEndpoint{d.Service, string(d.Kind), d.Key, d.Display})
	}
	for _, o := range res.Orphans {
		ov.Orphans = append(ov.Orphans, OEndpoint{o.Service, string(o.Kind), o.Key, o.Display})
	}
	return ov
}

// SaveOverlay writes overlay.json deterministically.
func SaveOverlay(root string, ov *Overlay) error {
	if err := os.MkdirAll(filepath.Join(root, SystemDir), 0o755); err != nil {
		return err
	}
	b, err := marshal(ov)
	if err != nil {
		return err
	}
	return os.WriteFile(overlayPath(root), b, 0o644)
}

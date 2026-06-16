package workspace

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/evgeniy-achin/graffiti/internal/graph"
	"github.com/evgeniy-achin/graffiti/internal/store"
)

// memberIndexes loads each member's map.json into a store.Index keyed by alias.
// Also returns the sha256 of each member's map.json (for source_hashes / staleness).
func memberIndexes(root string, reg *Registry) (map[string]*store.Index, map[string]string, error) {
	idxByAlias := make(map[string]*store.Index, len(reg.Members))
	hashes := make(map[string]string, len(reg.Members))
	for _, m := range reg.Members {
		dir := filepath.Join(root, filepath.FromSlash(m.Path))
		doc, err := store.Load(filepath.Join(dir, ".graffiti", "map.json"))
		if err != nil {
			return nil, nil, fmt.Errorf("workspace: member %q: %w", m.Alias, err)
		}
		idxByAlias[m.Alias] = store.NewIndex(doc)
		h, err := MapHash(dir)
		if err != nil {
			return nil, nil, err
		}
		hashes[m.Alias] = h
	}
	return idxByAlias, hashes, nil
}

// ComputeOverlay resolves explicit links against the members' current map.json
// files (signal 1). A link whose BOTH endpoints resolve to real nodes becomes an
// EXTRACTED cross-edge (via: explicit); otherwise it is recorded in Unresolved
// (never emitted as a confident edge — honesty-first under-linking, §16.2).
func ComputeOverlay(root string, reg *Registry, links []ParsedLink) (*Overlay, error) {
	idxByAlias, hashes, err := memberIndexes(root, reg)
	if err != nil {
		return nil, err
	}
	ov := &Overlay{Version: SchemaVersion, SourceHashes: hashes}
	resolves := func(alias, id string) bool {
		idx, ok := idxByAlias[alias]
		if !ok {
			return false
		}
		_, ok = idx.Node(id)
		return ok
	}
	for _, pl := range links {
		l := Link{
			From: pl.FromAlias + "::" + pl.FromID,
			To:   pl.ToAlias + "::" + pl.ToID,
			Relation: pl.Relation, Confidence: string(graph.ConfExtracted), Via: "explicit",
		}
		if resolves(pl.FromAlias, pl.FromID) && resolves(pl.ToAlias, pl.ToID) {
			ov.Links = append(ov.Links, l)
		} else {
			ov.Unresolved = append(ov.Unresolved, l)
		}
	}
	sortLinks(ov.Links)
	sortLinks(ov.Unresolved)
	return ov, nil
}

func sortLinks(ls []Link) {
	sort.SliceStable(ls, func(i, j int) bool {
		a, b := ls[i], ls[j]
		if a.From != b.From {
			return a.From < b.From
		}
		if a.To != b.To {
			return a.To < b.To
		}
		return a.Relation < b.Relation
	})
}

func overlayPath(root string) string { return filepath.Join(root, WorkspaceDir, overlayFile) }

// SaveOverlay writes overlay.json (links sorted, 2-space indent, trailing newline).
func SaveOverlay(root string, ov *Overlay) error {
	sortLinks(ov.Links)
	sortLinks(ov.Ambiguous)
	sortLinks(ov.Unresolved)
	if err := os.MkdirAll(filepath.Join(root, WorkspaceDir), 0o755); err != nil {
		return err
	}
	b, err := json.MarshalIndent(ov, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(overlayPath(root), append(b, '\n'), 0o644)
}

// LoadOverlay reads overlay.json with links sorted.
func LoadOverlay(root string) (*Overlay, error) {
	b, err := os.ReadFile(overlayPath(root))
	if err != nil {
		return nil, fmt.Errorf("workspace: read overlay: %w", err)
	}
	var ov Overlay
	if err := json.Unmarshal(b, &ov); err != nil {
		return nil, fmt.Errorf("workspace: parse overlay: %w", err)
	}
	sortLinks(ov.Links)
	return &ov, nil
}

// StaleMembers returns the aliases whose current map.json hash differs from the
// hash the overlay was computed against (self-healing nudge, §16.3).
func StaleMembers(root string, reg *Registry, ov *Overlay) ([]string, error) {
	_, hashes, err := memberIndexes(root, reg)
	if err != nil {
		return nil, err
	}
	var stale []string
	for alias, cur := range hashes {
		if ov.SourceHashes[alias] != cur {
			stale = append(stale, alias)
		}
	}
	sort.Strings(stale)
	return stale, nil
}

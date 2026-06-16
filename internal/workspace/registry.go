package workspace

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

// MapHash returns the lowercase-hex sha256 of <memberDir>/.graffiti/map.json.
func MapHash(memberDir string) (string, error) {
	b, err := os.ReadFile(filepath.Join(memberDir, ".graffiti", "map.json"))
	if err != nil {
		return "", fmt.Errorf("workspace: read member map.json: %w", err)
	}
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:]), nil
}

// AddMember inserts or replaces (by alias) a member, keeping Members sorted by alias.
func AddMember(reg *Registry, m Member) {
	for i := range reg.Members {
		if reg.Members[i].Alias == m.Alias {
			reg.Members[i] = m
			sortMembers(reg)
			return
		}
	}
	reg.Members = append(reg.Members, m)
	sortMembers(reg)
}

// RemoveMember drops the member with the given alias; reports whether one was removed.
func RemoveMember(reg *Registry, alias string) bool {
	out := reg.Members[:0]
	removed := false
	for _, m := range reg.Members {
		if m.Alias == alias {
			removed = true
			continue
		}
		out = append(out, m)
	}
	reg.Members = out
	return removed
}

func sortMembers(reg *Registry) {
	sort.SliceStable(reg.Members, func(i, j int) bool { return reg.Members[i].Alias < reg.Members[j].Alias })
}

func registryPath(root string) string { return filepath.Join(root, WorkspaceDir, registryFile) }

// SaveRegistry writes workspace.json (members sorted, 2-space indent, trailing newline).
func SaveRegistry(root string, reg *Registry) error {
	sortMembers(reg)
	if err := os.MkdirAll(filepath.Join(root, WorkspaceDir), 0o755); err != nil {
		return err
	}
	b, err := json.MarshalIndent(reg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(registryPath(root), append(b, '\n'), 0o644)
}

// LoadRegistry reads workspace.json and returns it with members sorted by alias.
func LoadRegistry(root string) (*Registry, error) {
	b, err := os.ReadFile(registryPath(root))
	if err != nil {
		return nil, fmt.Errorf("workspace: read registry: %w", err)
	}
	var reg Registry
	if err := json.Unmarshal(b, &reg); err != nil {
		return nil, fmt.Errorf("workspace: parse registry: %w", err)
	}
	sortMembers(&reg)
	return &reg, nil
}

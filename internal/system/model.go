// Package system orchestrates many independent service repos into ONE system
// graph. Each service publishes its own map.json (with a contract surface) into a
// shared store (git-as-registry); the orchestrator collects the published
// artifacts, federates them, and DISCOVERS cross-service edges by matching what
// each service consumes against what others provide. It reuses the read-side graph
// model and the Plan-9 viewer; it adds no daemon — orchestration is CLI + a
// committable registry + a derived, recomputable overlay.
package system

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

// SystemDir is the per-system metadata directory (sibling to .graffiti).
const SystemDir = ".graffiti-system"

const (
	registryFile = "system.json"
	overlayFile  = "overlay.json"
	servicesDir  = "services"
	// Version is the system.json schema version.
	Version = "1"
)

// Service is one published member of the system.
type Service struct {
	Name     string `json:"name"`
	Path     string `json:"path"`               // artifact dir, relative to system root
	Language string `json:"language,omitempty"`
	Owner    string `json:"owner,omitempty"`
	Commit   string `json:"commit,omitempty"`   // pinned source SHA at publish time
	MapHash  string `json:"map_hash,omitempty"` // hash of the published map.json
}

// Registry is the committable system manifest (.graffiti-system/system.json).
type Registry struct {
	Version     string    `json:"version"`
	Name        string    `json:"name"`
	GeneratedAt string    `json:"generated_at"`
	Services    []Service `json:"services"`
}

// NewRegistry returns an empty registry.
func NewRegistry(name string) *Registry {
	return &Registry{Version: Version, Name: name, Services: []Service{}}
}

// AddService inserts or replaces a service by name, keeping the slice sorted.
func AddService(reg *Registry, s Service) {
	for i := range reg.Services {
		if reg.Services[i].Name == s.Name {
			reg.Services[i] = s
			sortServices(reg)
			return
		}
	}
	reg.Services = append(reg.Services, s)
	sortServices(reg)
}

// RemoveService drops a service by name; reports whether it existed.
func RemoveService(reg *Registry, name string) bool {
	for i := range reg.Services {
		if reg.Services[i].Name == name {
			reg.Services = append(reg.Services[:i], reg.Services[i+1:]...)
			return true
		}
	}
	return false
}

func sortServices(reg *Registry) {
	sort.Slice(reg.Services, func(i, j int) bool { return reg.Services[i].Name < reg.Services[j].Name })
}

func registryPath(root string) string { return filepath.Join(root, SystemDir, registryFile) }
func overlayPath(root string) string  { return filepath.Join(root, SystemDir, overlayFile) }

// artifactMapPath is where a service's published map.json lives.
func artifactMapPath(root, name string) string {
	return filepath.Join(root, SystemDir, servicesDir, name, "map.json")
}

// SaveRegistry writes system.json deterministically (sorted services, 2-space).
func SaveRegistry(root string, reg *Registry) error {
	sortServices(reg)
	if err := os.MkdirAll(filepath.Join(root, SystemDir), 0o755); err != nil {
		return err
	}
	b, err := marshal(reg)
	if err != nil {
		return err
	}
	return os.WriteFile(registryPath(root), b, 0o644)
}

// LoadRegistry reads system.json.
func LoadRegistry(root string) (*Registry, error) {
	b, err := os.ReadFile(registryPath(root))
	if err != nil {
		return nil, fmt.Errorf("system: read registry: %w", err)
	}
	var reg Registry
	if err := json.Unmarshal(b, &reg); err != nil {
		return nil, fmt.Errorf("system: parse registry: %w", err)
	}
	return &reg, nil
}

// FindRoot walks up from cwd to find a directory containing .graffiti-system.
func FindRoot(cwd string) string {
	dir := cwd
	for {
		if fi, err := os.Stat(filepath.Join(dir, SystemDir, registryFile)); err == nil && !fi.IsDir() {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return ""
		}
		dir = parent
	}
}

func marshal(v any) ([]byte, error) {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return nil, err
	}
	return append(b, '\n'), nil
}

func hashBytes(b []byte) string {
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])[:12]
}

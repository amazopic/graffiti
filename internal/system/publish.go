package system

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Publish copies a service's already-built map.json into the system store at
// <systemRoot>/.graffiti-system/services/<name>/map.json and returns the Service
// metadata (pinned commit + artifact hash) to record in the registry. The service
// must have been built (graffiti build) first.
func Publish(serviceRoot, systemRoot, name string) (Service, error) {
	src := filepath.Join(serviceRoot, ".graffiti", "map.json")
	b, err := os.ReadFile(src)
	if err != nil {
		return Service{}, fmt.Errorf("system: read %s (run `graffiti build` in the service first): %w", src, err)
	}
	destDir := filepath.Join(systemRoot, SystemDir, servicesDir, name)
	if err := os.MkdirAll(destDir, 0o755); err != nil {
		return Service{}, err
	}
	if err := os.WriteFile(filepath.Join(destDir, "map.json"), b, 0o644); err != nil {
		return Service{}, err
	}
	return Service{
		Name:    name,
		Path:    filepath.ToSlash(filepath.Join(SystemDir, servicesDir, name)),
		Commit:  gitCommit(serviceRoot),
		MapHash: hashBytes(b),
	}, nil
}

// gitCommit returns the short HEAD SHA of dir, or "" if not a git repo / no git.
func gitCommit(dir string) string {
	cmd := exec.Command("git", "-C", dir, "rev-parse", "--short", "HEAD")
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

// StaleServices returns the names of services whose on-disk published artifact no
// longer matches the hash recorded in the registry (republished but not
// re-federated, or missing).
func StaleServices(root string, reg *Registry) []string {
	var stale []string
	for _, s := range reg.Services {
		b, err := os.ReadFile(artifactMapPath(root, s.Name))
		if err != nil {
			stale = append(stale, s.Name+" (missing)")
			continue
		}
		if s.MapHash != "" && hashBytes(b) != s.MapHash {
			stale = append(stale, s.Name+" (changed)")
		}
	}
	return stale
}

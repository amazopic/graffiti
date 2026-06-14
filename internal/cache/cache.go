// Package cache stores per-file SHA256 content hashes under .graffiti/cache/
// to support incremental rebuilds (spec §6). Plan 1 only writes/reads hashes;
// it performs no skip-on-match (that is a later plan).
package cache

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
)

// HashBytes returns the lowercase hex SHA256 of b.
func HashBytes(b []byte) string {
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}

// HashFile returns the SHA256 of the file at path.
func HashFile(path string) (string, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return HashBytes(b), nil
}

// Cache holds repo-relative path -> content hash, persisted to
// <root>/.graffiti/cache/hashes.json.
type Cache struct {
	root    string
	entries map[string]string
}

// New returns an empty cache rooted at repo root.
func New(root string) *Cache {
	return &Cache{root: root, entries: map[string]string{}}
}

func (c *Cache) path() string {
	return filepath.Join(c.root, ".graffiti", "cache", "hashes.json")
}

// Load reads existing cache entries; a missing file is not an error.
func (c *Cache) Load() error {
	b, err := os.ReadFile(c.path())
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	return json.Unmarshal(b, &c.entries)
}

// Get returns the stored hash for relPath.
func (c *Cache) Get(relPath string) (string, bool) {
	h, ok := c.entries[relPath]
	return h, ok
}

// Put records a hash for relPath.
func (c *Cache) Put(relPath, hash string) error {
	c.entries[relPath] = hash
	return nil
}

// Flush writes the cache deterministically (sorted keys) to disk.
func (c *Cache) Flush() error {
	if err := os.MkdirAll(filepath.Dir(c.path()), 0o755); err != nil {
		return err
	}
	keys := make([]string, 0, len(c.entries))
	for k := range c.entries {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	ordered := make([][2]string, 0, len(keys))
	for _, k := range keys {
		ordered = append(ordered, [2]string{k, c.entries[k]})
	}
	b, err := marshalSorted(ordered)
	if err != nil {
		return err
	}
	return os.WriteFile(c.path(), b, 0o644)
}

// marshalSorted encodes [(key,value)...] as a JSON object in the given order.
func marshalSorted(pairs [][2]string) ([]byte, error) {
	out := []byte("{")
	for i, p := range pairs {
		if i > 0 {
			out = append(out, ',')
		}
		kb, err := json.Marshal(p[0])
		if err != nil {
			return nil, err
		}
		vb, err := json.Marshal(p[1])
		if err != nil {
			return nil, err
		}
		out = append(out, kb...)
		out = append(out, ':')
		out = append(out, vb...)
	}
	out = append(out, '}', '\n')
	return out, nil
}

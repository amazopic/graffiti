package cache

import (
	"os"
	"path/filepath"
	"testing"
)

func TestHashBytes_Stable(t *testing.T) {
	a := HashBytes([]byte("hello"))
	b := HashBytes([]byte("hello"))
	if a != b {
		t.Fatalf("hash not stable: %q vs %q", a, b)
	}
	const want = "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824"
	if a != want {
		t.Fatalf("hash = %q, want %q", a, want)
	}
}

func TestHashFile(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "x.go")
	if err := os.WriteFile(p, []byte("package main"), 0o644); err != nil {
		t.Fatal(err)
	}
	h, err := HashFile(p)
	if err != nil {
		t.Fatalf("HashFile: %v", err)
	}
	if h != HashBytes([]byte("package main")) {
		t.Fatalf("HashFile mismatch")
	}
}

func TestCacheRoundTrip(t *testing.T) {
	dir := t.TempDir()
	c := New(dir)
	if err := c.Put("main.go", "abc123"); err != nil {
		t.Fatalf("Put: %v", err)
	}
	if err := c.Flush(); err != nil {
		t.Fatalf("Flush: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, ".graffiti", "cache", "hashes.json")); err != nil {
		t.Fatalf("cache file missing: %v", err)
	}
	c2 := New(dir)
	if err := c2.Load(); err != nil {
		t.Fatalf("Load: %v", err)
	}
	if got, ok := c2.Get("main.go"); !ok || got != "abc123" {
		t.Fatalf("reload mismatch: ok=%v got=%q", ok, got)
	}
}

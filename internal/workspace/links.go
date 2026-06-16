package workspace

import (
	"fmt"
	"strings"
)

// ParsedLink is one parsed explicit link (signal 1, §16.2). Endpoints are split
// into alias + local node id; the alias-qualified id is "alias::id".
type ParsedLink struct {
	FromAlias, FromID string
	ToAlias, ToID     string
	Relation          string
}

// validRelations mirrors graph.ValidRelations (kept local to avoid importing the
// graph package here; the values are the §6 relation vocabulary).
var validRelations = map[string]bool{
	"calls": true, "imports": true, "inherits": true,
	"implements": true, "references": true, "contains": true,
}

// ParseLinks parses the line-based explicit-links file (§16.2 signal 1, v1
// format): one link per non-blank, non-comment line:
//
//	FROM -> TO [relation]      # FROM and TO are "alias::nodeid"
//
// Default relation is "references". '#' starts a comment (to end of line).
func ParseLinks(b []byte) ([]ParsedLink, error) {
	var out []ParsedLink
	for i, raw := range strings.Split(string(b), "\n") {
		line := raw
		if h := strings.IndexByte(line, '#'); h >= 0 {
			line = line[:h]
		}
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		lhs, rhs, ok := strings.Cut(line, "->")
		if !ok {
			return nil, fmt.Errorf("links line %d: missing '->': %q", i+1, raw)
		}
		fromA, fromID, err := splitEndpoint(strings.TrimSpace(lhs))
		if err != nil {
			return nil, fmt.Errorf("links line %d: from: %w", i+1, err)
		}
		fields := strings.Fields(strings.TrimSpace(rhs))
		if len(fields) == 0 {
			return nil, fmt.Errorf("links line %d: missing target", i+1)
		}
		toA, toID, err := splitEndpoint(fields[0])
		if err != nil {
			return nil, fmt.Errorf("links line %d: to: %w", i+1, err)
		}
		rel := "references"
		if len(fields) >= 2 {
			rel = fields[1]
			if !validRelations[rel] {
				return nil, fmt.Errorf("links line %d: unknown relation %q", i+1, rel)
			}
		}
		out = append(out, ParsedLink{fromA, fromID, toA, toID, rel})
	}
	return out, nil
}

func splitEndpoint(s string) (alias, id string, err error) {
	a, i, ok := strings.Cut(s, "::")
	if !ok {
		return "", "", fmt.Errorf("endpoint %q is not alias::id", s)
	}
	if a == "" || i == "" {
		return "", "", fmt.Errorf("endpoint %q has empty alias or id", s)
	}
	return a, i, nil
}

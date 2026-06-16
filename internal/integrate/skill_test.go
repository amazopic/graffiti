package integrate

import (
	"strings"
	"testing"
)

func TestSkillContent_Shape(t *testing.T) {
	s := SkillContent()
	if !strings.HasPrefix(s, "---\nname: graffiti\n") {
		head := s
		if len(head) > 40 {
			head = head[:40]
		}
		t.Fatalf("skill must start with YAML frontmatter incl. name: graffiti\n%q", head)
	}
	if !strings.Contains(s, "description:") {
		t.Fatal("skill frontmatter missing description")
	}
	// Frontmatter closes before the body heading.
	if strings.Count(s, "\n---\n") < 1 {
		t.Fatal("skill frontmatter not closed with ---")
	}
	for _, must := range []string{
		"graffiti build .",
		"graffiti query",
		"graffiti update",
		".graffiti/MAP.md",
	} {
		if !strings.Contains(s, must) {
			t.Fatalf("skill missing %q", must)
		}
	}
	if !strings.HasSuffix(s, "\n") {
		t.Fatal("skill must end with a trailing newline")
	}
	// description stays within Claude Code's 1536-char cap.
	desc := s[strings.Index(s, "description:"):]
	desc = desc[:strings.Index(desc, "\n")]
	if len(desc) > 1536 {
		t.Fatalf("description too long: %d chars", len(desc))
	}
}

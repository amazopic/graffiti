package integrate

import (
	"path/filepath"
	"testing"
)

func TestProjectTarget_Paths(t *testing.T) {
	tg := ProjectTarget("/repo")
	if tg.Scope != "project" {
		t.Fatalf("scope = %q", tg.Scope)
	}
	if tg.SkillPath != filepath.FromSlash("/repo/.claude/skills/graffiti/SKILL.md") {
		t.Fatalf("skill path = %q", tg.SkillPath)
	}
	if tg.ClaudeMDPath != filepath.FromSlash("/repo/CLAUDE.md") {
		t.Fatalf("claude md path = %q", tg.ClaudeMDPath)
	}
	if tg.SettingsPath != filepath.FromSlash("/repo/.claude/settings.json") {
		t.Fatalf("settings path = %q", tg.SettingsPath)
	}
}

func TestUserTarget_Paths(t *testing.T) {
	tg := UserTarget("/home/u")
	if tg.Scope != "user" {
		t.Fatalf("scope = %q", tg.Scope)
	}
	if tg.SkillPath != filepath.FromSlash("/home/u/.claude/skills/graffiti/SKILL.md") {
		t.Fatalf("skill path = %q", tg.SkillPath)
	}
	// User-scoped CLAUDE.md lives under ~/.claude, not the home root.
	if tg.ClaudeMDPath != filepath.FromSlash("/home/u/.claude/CLAUDE.md") {
		t.Fatalf("claude md path = %q", tg.ClaudeMDPath)
	}
	if tg.SettingsPath != filepath.FromSlash("/home/u/.claude/settings.json") {
		t.Fatalf("settings path = %q", tg.SettingsPath)
	}
}

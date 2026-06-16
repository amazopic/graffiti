// Package integrate generates and installs graffiti's Claude Code integration:
// a Skill, an always-on CLAUDE.md block, and an optional PreToolUse hook (spec §9).
// All generated content is constant (byte-deterministic); merges are idempotent.
package integrate

import "path/filepath"

// Target is the set of absolute destination paths for one install scope.
// The CLI computes these from --user/--hook flags and hands them to Install,
// which keeps Install free of any environment/home-dir coupling (fully testable).
type Target struct {
	Scope        string // "project" or "user" — for the success message only
	SkillPath    string // .../.claude/skills/graffiti/SKILL.md
	ClaudeMDPath string // CLAUDE.md (repo root for project, ~/.claude for user)
	SettingsPath string // .../.claude/settings.json
}

// ProjectTarget installs into a repository rooted at root. CLAUDE.md sits at the
// repo root (the conventional project memory file).
func ProjectTarget(root string) Target {
	claudeDir := filepath.Join(root, ".claude")
	return Target{
		Scope:        "project",
		SkillPath:    filepath.Join(claudeDir, "skills", "graffiti", "SKILL.md"),
		ClaudeMDPath: filepath.Join(root, "CLAUDE.md"),
		SettingsPath: filepath.Join(claudeDir, "settings.json"),
	}
}

// UserTarget installs into the user's home. CLAUDE.md sits under ~/.claude.
func UserTarget(home string) Target {
	claudeDir := filepath.Join(home, ".claude")
	return Target{
		Scope:        "user",
		SkillPath:    filepath.Join(claudeDir, "skills", "graffiti", "SKILL.md"),
		ClaudeMDPath: filepath.Join(claudeDir, "CLAUDE.md"),
		SettingsPath: filepath.Join(claudeDir, "settings.json"),
	}
}

package integrate

import (
	"os"
	"testing"
)

func readFile(t *testing.T, p string) string {
	t.Helper()
	b, err := os.ReadFile(p)
	if err != nil {
		t.Fatalf("read %s: %v", p, err)
	}
	return string(b)
}

func TestInstall_ProjectNoHook(t *testing.T) {
	root := t.TempDir()
	tg := ProjectTarget(root)
	res, err := Install(tg, Options{InstallHook: false})
	if err != nil {
		t.Fatal(err)
	}
	if res.Skill != ActionCreated {
		t.Fatalf("skill action = %v, want Created", res.Skill)
	}
	if res.ClaudeMD != ActionCreated {
		t.Fatalf("claudemd action = %v, want Created", res.ClaudeMD)
	}
	if res.HookInstalled {
		t.Fatal("hook must not be installed without InstallHook")
	}
	// Files exist with expected content.
	if got := readFile(t, tg.SkillPath); got != SkillContent() {
		t.Fatal("SKILL.md content mismatch")
	}
	if got := readFile(t, tg.ClaudeMDPath); got != ClaudeBlock() {
		t.Fatal("CLAUDE.md content mismatch")
	}
	if _, err := os.Stat(tg.SettingsPath); !os.IsNotExist(err) {
		t.Fatal("settings.json must not be written without --hook")
	}
}

func TestInstall_ProjectWithHook(t *testing.T) {
	root := t.TempDir()
	tg := ProjectTarget(root)
	res, err := Install(tg, Options{InstallHook: true})
	if err != nil {
		t.Fatal(err)
	}
	if !res.HookInstalled || res.Hook != ActionCreated {
		t.Fatalf("hook action = %v installed=%v", res.Hook, res.HookInstalled)
	}
	got := readFile(t, tg.SettingsPath)
	if !containsAll(got, "PreToolUse", HookCommand) {
		t.Fatalf("settings.json missing hook:\n%s", got)
	}
}

func TestInstall_IdempotentSecondRunUnchanged(t *testing.T) {
	root := t.TempDir()
	tg := ProjectTarget(root)
	if _, err := Install(tg, Options{InstallHook: true}); err != nil {
		t.Fatal(err)
	}
	res, err := Install(tg, Options{InstallHook: true})
	if err != nil {
		t.Fatal(err)
	}
	if res.Skill != ActionUnchanged || res.ClaudeMD != ActionUnchanged || res.Hook != ActionUnchanged {
		t.Fatalf("second run should be all-Unchanged, got skill=%v claude=%v hook=%v",
			res.Skill, res.ClaudeMD, res.Hook)
	}
}

func TestInstall_PreservesExistingClaudeMD(t *testing.T) {
	root := t.TempDir()
	tg := ProjectTarget(root)
	original := "# House rules\n\nAlways write tests.\n"
	if err := os.WriteFile(tg.ClaudeMDPath, []byte(original), 0o644); err != nil {
		t.Fatal(err)
	}
	res, err := Install(tg, Options{})
	if err != nil {
		t.Fatal(err)
	}
	if res.ClaudeMD != ActionUpdated {
		t.Fatalf("claudemd action = %v, want Updated", res.ClaudeMD)
	}
	got := readFile(t, tg.ClaudeMDPath)
	if !containsAll(got, "House rules", "Always write tests.", claudeStart) {
		t.Fatalf("existing CLAUDE.md content not preserved:\n%s", got)
	}
}

func TestInstall_UserScopeIntoTempHome(t *testing.T) {
	home := t.TempDir()
	tg := UserTarget(home)
	if _, err := Install(tg, Options{InstallHook: true}); err != nil {
		t.Fatal(err)
	}
	for _, p := range []string{tg.SkillPath, tg.ClaudeMDPath, tg.SettingsPath} {
		if _, err := os.Stat(p); err != nil {
			t.Fatalf("expected %s to exist: %v", p, err)
		}
	}
}

func containsAll(s string, subs ...string) bool {
	for _, sub := range subs {
		found := false
		for i := 0; i+len(sub) <= len(s); i++ {
			if s[i:i+len(sub)] == sub {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

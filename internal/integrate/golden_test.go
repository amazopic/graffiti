package integrate

import (
	"os"
	"path/filepath"
	"testing"
)

// goldenDir resolves testdata/golden/init relative to the repo root (two levels
// up from internal/integrate).
func goldenDir(t *testing.T) string {
	t.Helper()
	return filepath.FromSlash("../../testdata/golden/init")
}

func TestInstall_MatchesGolden(t *testing.T) {
	root := t.TempDir()
	tg := ProjectTarget(root)
	if _, err := Install(tg, Options{InstallHook: true}); err != nil {
		t.Fatal(err)
	}
	cases := []struct{ got, golden string }{
		{tg.SkillPath, "SKILL.md"},
		{tg.ClaudeMDPath, "CLAUDE.md"},
		{tg.SettingsPath, "settings.json"},
	}
	for _, c := range cases {
		gotB, err := os.ReadFile(c.got)
		if err != nil {
			t.Fatalf("read produced %s: %v", c.golden, err)
		}
		wantB, err := os.ReadFile(filepath.Join(goldenDir(t), c.golden))
		if err != nil {
			t.Fatalf("read golden %s: %v", c.golden, err)
		}
		if string(gotB) != string(wantB) {
			t.Fatalf("%s differs from golden.\n--- got ---\n%s\n--- want ---\n%s", c.golden, gotB, wantB)
		}
	}
}

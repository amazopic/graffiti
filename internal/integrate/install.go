package integrate

import (
	"os"
	"path/filepath"
)

// Options controls what Install writes.
type Options struct {
	InstallHook bool // also merge the optional PreToolUse hook into settings.json
}

// Action records what happened to one artifact.
type Action int

const (
	ActionUnchanged Action = iota // already up to date
	ActionCreated                 // file did not exist; created
	ActionUpdated                 // file existed; content changed
)

func (a Action) String() string {
	switch a {
	case ActionCreated:
		return "created"
	case ActionUpdated:
		return "updated"
	default:
		return "unchanged"
	}
}

// Result summarizes one Install run for the CLI success message.
type Result struct {
	Target        Target
	Skill         Action
	ClaudeMD      Action
	Hook          Action
	HookInstalled bool // whether the hook was part of this run (Options.InstallHook)
}

// Install writes the three integration artifacts to the target paths. It creates
// parent directories as needed and is idempotent — a second run reports Unchanged.
func Install(t Target, opts Options) (Result, error) {
	res := Result{Target: t, HookInstalled: opts.InstallHook}

	// 1. Skill — we own this path; desired content is constant.
	skillAct, err := writeIfChanged(t.SkillPath, []byte(SkillContent()))
	if err != nil {
		return res, err
	}
	res.Skill = skillAct

	// 2. CLAUDE.md — surgical merge preserving user content.
	existingMD, err := readMaybe(t.ClaudeMDPath)
	if err != nil {
		return res, err
	}
	mergedMD := MergeClaudeMD(existingMD)
	mdAct, err := writeIfChanged(t.ClaudeMDPath, mergedMD)
	if err != nil {
		return res, err
	}
	res.ClaudeMD = mdAct

	// 3. settings.json hook — only with --hook.
	if opts.InstallHook {
		existingSettings, err := readMaybe(t.SettingsPath)
		if err != nil {
			return res, err
		}
		mergedSettings, err := MergeHookSettings(existingSettings)
		if err != nil {
			return res, err
		}
		hookAct, err := writeIfChanged(t.SettingsPath, mergedSettings)
		if err != nil {
			return res, err
		}
		res.Hook = hookAct
	}

	return res, nil
}

// readMaybe reads a file, treating "not found" as empty (no error).
func readMaybe(path string) ([]byte, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	return b, nil
}

// writeIfChanged writes content to path only if it differs from what's there,
// creating parent dirs. It returns Created (new file), Updated (changed), or
// Unchanged (byte-identical).
func writeIfChanged(path string, content []byte) (Action, error) {
	existing, err := os.ReadFile(path)
	switch {
	case err == nil:
		if string(existing) == string(content) {
			return ActionUnchanged, nil
		}
	case !os.IsNotExist(err):
		return ActionUnchanged, err
	}
	created := os.IsNotExist(err)
	if mkErr := os.MkdirAll(filepath.Dir(path), 0o755); mkErr != nil {
		return ActionUnchanged, mkErr
	}
	if wErr := os.WriteFile(path, content, 0o644); wErr != nil {
		return ActionUnchanged, wErr
	}
	if created {
		return ActionCreated, nil
	}
	return ActionUpdated, nil
}

package integrate

import "bytes"

const (
	claudeStart = "<!-- graffiti:start -->"
	claudeEnd   = "<!-- graffiti:end -->"
)

// ClaudeBlock is the always-on CLAUDE.md block (spec §9), wrapped in HTML-comment
// markers so it can be refreshed in place on re-run.
func ClaudeBlock() string {
	return claudeStart + `
## graffiti code map

If ` + "`.graffiti/map.json`" + ` exists, this repo has a graffiti code map. For questions about the
codebase's structure (where something lives, how parts connect, the architecture), run
` + "`graffiti query \"<question>\"`" + ` instead of grep/read — it returns a scoped subgraph. After
editing code, run ` + "`graffiti update`" + ` to refresh the map. If no map exists yet, run
` + "`graffiti build .`" + ` first.
` + claudeEnd + "\n"
}

// MergeClaudeMD returns the new CLAUDE.md content after inserting or refreshing
// the graffiti block. If both markers are present, the content between them is
// replaced (allowing content upgrades); otherwise the block is appended after a
// blank-line separator. Everything outside the markers is preserved byte-for-byte.
// The result is idempotent: MergeClaudeMD(MergeClaudeMD(x)) == MergeClaudeMD(x).
func MergeClaudeMD(existing []byte) []byte {
	block := ClaudeBlock()

	if i := bytes.Index(existing, []byte(claudeStart)); i >= 0 {
		if j := bytes.Index(existing[i:], []byte(claudeEnd)); j >= 0 {
			end := i + j + len(claudeEnd)
			// Absorb a single newline right after the end marker so the block's own
			// trailing newline doesn't accumulate blank lines across re-runs.
			if end < len(existing) && existing[end] == '\n' {
				end++
			}
			var out bytes.Buffer
			out.Write(existing[:i])
			out.WriteString(block)
			out.Write(existing[end:])
			return out.Bytes()
		}
		// start marker without a matching end marker: malformed; fall through to append.
	}

	if len(existing) == 0 {
		return []byte(block)
	}
	var out bytes.Buffer
	out.Write(existing)
	if !bytes.HasSuffix(existing, []byte("\n")) {
		out.WriteByte('\n')
	}
	out.WriteByte('\n')
	out.WriteString(block)
	return out.Bytes()
}

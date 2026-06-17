// Command graffiti turns a code repository into a queryable directed knowledge graph.
package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/amazopic/graffiti/internal/app"
	"github.com/amazopic/graffiti/internal/integrate"
	"github.com/amazopic/graffiti/internal/mcp"
	"github.com/amazopic/graffiti/internal/query"
	"github.com/amazopic/graffiti/internal/render"
	"github.com/amazopic/graffiti/internal/store"
	"github.com/amazopic/graffiti/internal/workspace"
)

// version is the build version, injected at release time via
// -ldflags "-X main.version=<tag>". Defaults to "dev" for local builds.
var version = "dev"

func main() {
	os.Exit(run(os.Args, os.Stdin, os.Stdout, os.Stderr))
}

// run is the testable entry point. It returns the process exit code.
func run(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	if len(args) < 2 {
		usage(stderr)
		return 2
	}

	cmd := args[1]
	switch cmd {
	case ".":
		return runBuild(".", stdout, stderr)
	case "build":
		root := "."
		if len(args) >= 3 {
			root = args[2]
		}
		return runBuild(root, stdout, stderr)
	case "update":
		if hasFlag(args[2:], "--workspace") {
			return updateWorkspace(stripFlag(args[2:], "--workspace"), stdout, stderr)
		}
		// update is currently a full rebuild; the incremental AST-only rebuild
		// (spec §11) is a later optimization. Behaves exactly like `build`.
		root := "."
		if len(args) >= 3 {
			root = args[2]
		}
		return runBuild(root, stdout, stderr)
	case "query":
		qargs := args[2:]
		if hasFlag(qargs, "--workspace") {
			return runQueryWorkspace(stripFlag(qargs, "--workspace"), stdout, stderr)
		}
		if len(args) < 3 {
			fmt.Fprintln(stderr, "graffiti: query requires a question")
			usage(stderr)
			return 2
		}
		// Guard: question + optional path = at most 2 extra positional args.
		// More than that means the user forgot to quote a multi-word question.
		if len(args) > 4 {
			fmt.Fprintln(stderr, `graffiti: too many arguments for query — did you forget to quote the question?`)
			fmt.Fprintln(stderr, `  example: graffiti query "login handler" [path]`)
			return 2
		}
		question := args[2]
		root := "."
		if len(args) >= 4 {
			root = args[3]
		}
		return runQuery(root, question, stdout, stderr)
	case "serve":
		sargs := args[2:]
		if hasFlag(sargs, "--workspace") {
			return serveWorkspace(stripFlag(sargs, "--workspace"), stdin, stdout, stderr)
		}
		root := "."
		if len(args) >= 3 {
			root = args[2]
		}
		return serve(root, stdin, stdout, stderr)
	case "link":
		return runLink(args[2:], stdout, stderr)
	case "workspace":
		return runWorkspace(args[2:], stdout, stderr)
	case "links":
		return runLinksCheck(args[2:], stdout, stderr)
	case "federate":
		return runFederateExplain(args[2:], stdout, stderr)
	case "init":
		return runInit(args[2:], stdout, stderr)
	case "hook":
		// Internal: PreToolUse handler. Always exits 0; never blocks a tool.
		cwd, _ := os.Getwd()
		integrate.RunHook(stdin, stdout, cwd)
		return 0
	case "version", "--version", "-v":
		fmt.Fprintln(stdout, "graffiti "+version)
		return 0
	default:
		// Treat an existing path as `build <path>` for the common `graffiti <path>` form.
		if info, err := os.Stat(cmd); err == nil && info.IsDir() {
			return runBuild(cmd, stdout, stderr)
		}
		fmt.Fprintf(stderr, "graffiti: unknown command %q\n", cmd)
		usage(stderr)
		return 2
	}
}

func usage(w io.Writer) {
	fmt.Fprintln(w, "usage: graffiti <command>")
	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "commands:")
	fmt.Fprintln(w, "  .                 build the map for the current repo")
	fmt.Fprintln(w, "  build <path>      build the map for <path> (default .)")
	fmt.Fprintln(w, "  update [path]     rebuild the map for <path> (default .)")
	fmt.Fprintln(w, `  query "<q>" [path]  LLM-free scoped subgraph retrieval (soft 2000-token node budget)`)
	fmt.Fprintln(w, "  serve [path]      MCP server over stdio (JSON-RPC 2.0)")
	fmt.Fprintln(w, "  init [--user] [--hook]  install Claude Code integration (skill + CLAUDE.md [+ hook])")
	fmt.Fprintln(w, "  link <pathA> <pathB> [...] [--name n]  federate projects into a workspace")
	fmt.Fprintln(w, "  workspace <add|rm|list|render> [--root dir]  manage workspace / render workspace.html")
	fmt.Fprintln(w, "  links check [--root dir]  validate explicit cross-project links resolve")
	fmt.Fprintln(w, "  federate --explain [--root dir]  print the computed cross-project links")
	fmt.Fprintln(w, `  query --workspace "<q>" [--root dir]  federated retrieval across the workspace`)
	fmt.Fprintln(w, "  serve --workspace [--root dir]  MCP server over the federated index")
	fmt.Fprintln(w, "  update --workspace [--root dir]  rebuild changed members + recompute links")
	fmt.Fprintln(w, "  version           print the graffiti version")
}

// runInit installs the Claude Code integration. Flags: --user (install into the
// home dir instead of the project), --hook (also install the PreToolUse hook),
// --root <dir> (project root; defaults to "." — primarily a test seam).
func runInit(args []string, stdout, stderr io.Writer) int {
	var user, hook bool
	root := "."
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--user":
			user = true
		case "--hook":
			hook = true
		case "--root":
			if i+1 >= len(args) {
				fmt.Fprintln(stderr, "graffiti: --root requires a directory")
				return 2
			}
			i++
			root = args[i]
		default:
			fmt.Fprintf(stderr, "graffiti: unknown init flag %q\n", args[i])
			return 2
		}
	}

	var target integrate.Target
	if user {
		home, err := os.UserHomeDir()
		if err != nil {
			fmt.Fprintf(stderr, "graffiti: cannot resolve home dir: %v\n", err)
			return 1
		}
		target = integrate.UserTarget(home)
	} else {
		target = integrate.ProjectTarget(root)
	}

	res, err := integrate.Install(target, integrate.Options{InstallHook: hook})
	if err != nil {
		fmt.Fprintf(stderr, "graffiti: init failed: %v\n", err)
		return 1
	}

	fmt.Fprintf(stdout, "✓ graffiti wired into Claude Code (%s).\n", res.Target.Scope)
	fmt.Fprintf(stdout, "  • skill:     %s (%s)\n", res.Target.SkillPath, res.Skill)
	fmt.Fprintf(stdout, "  • CLAUDE.md: %s (%s)\n", res.Target.ClaudeMDPath, res.ClaudeMD)
	if res.HookInstalled {
		fmt.Fprintf(stdout, "  • hook:      %s (%s) — PreToolUse nudge grep → graffiti query\n", res.Target.SettingsPath, res.Hook)
	} else {
		fmt.Fprintln(stdout, "  • hook:      skipped (pass --hook to nudge grep → graffiti query)")
	}
	fmt.Fprintln(stdout, "  Re-run `graffiti init` any time; it is idempotent.")
	return 0
}

// runLink builds any unbuilt members, auto-discovers the workspace root (nearest
// common ancestor), writes workspace.json, computes the overlay from links, and
// prints the success line. Flags: --name <name>.
func runLink(args []string, stdout, stderr io.Writer) int {
	name := "workspace"
	var paths []string
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--name":
			if i+1 >= len(args) {
				fmt.Fprintln(stderr, "graffiti: --name requires a value")
				return 2
			}
			i++
			name = args[i]
		default:
			paths = append(paths, args[i])
		}
	}
	if len(paths) < 2 {
		fmt.Fprintln(stderr, "graffiti: link requires at least two project paths")
		return 2
	}

	abs := make([]string, len(paths))
	for i, p := range paths {
		a, err := filepath.Abs(p)
		if err != nil {
			fmt.Fprintf(stderr, "graffiti: %v\n", err)
			return 1
		}
		abs[i] = a
	}
	root := workspace.CommonAncestor(abs)

	reg := &workspace.Registry{Version: workspace.SchemaVersion, Name: name, GeneratedAt: nowRFC3339()}
	for _, a := range abs {
		// build the member if it has no map.json yet
		if _, err := os.Stat(filepath.Join(a, ".graffiti", "map.json")); err != nil {
			if _, berr := app.Build(a, nowRFC3339()); berr != nil {
				fmt.Fprintf(stderr, "graffiti: build %s: %v\n", a, berr)
				return 1
			}
		}
		rel, err := filepath.Rel(root, a)
		if err != nil {
			fmt.Fprintf(stderr, "graffiti: %v\n", err)
			return 1
		}
		h, err := workspace.MapHash(a)
		if err != nil {
			fmt.Fprintf(stderr, "graffiti: %v\n", err)
			return 1
		}
		workspace.AddMember(reg, workspace.Member{Alias: aliasFor(a), Path: filepath.ToSlash(rel), MapHash: h})
	}
	if err := workspace.SaveRegistry(root, reg); err != nil {
		fmt.Fprintf(stderr, "graffiti: %v\n", err)
		return 1
	}
	ov, err := computeAndSaveOverlay(root, reg)
	if err != nil {
		fmt.Fprintf(stderr, "graffiti: %v\n", err)
		return 1
	}
	fmt.Fprintf(stdout, "✓ Linked %d projects. %d cross-project links (%d EXTRACTED, %d unresolved). 0 API calls, $0.\n",
		len(reg.Members), len(ov.Links), len(ov.Links), len(ov.Unresolved))
	return 0
}

// aliasFor derives a member alias from its directory base name.
func aliasFor(absPath string) string { return filepath.Base(absPath) }

// nowRFC3339 is the build/link timestamp (UTC, RFC3339).
func nowRFC3339() string { return time.Now().UTC().Format(time.RFC3339) }

// computeAndSaveOverlay reads <root>/.graffiti-workspace/links (if any), computes
// the overlay against the registry's members, stamps generated_at, and saves it.
func computeAndSaveOverlay(root string, reg *workspace.Registry) (*workspace.Overlay, error) {
	var links []workspace.ParsedLink
	if b, err := os.ReadFile(filepath.Join(root, workspace.WorkspaceDir, "links")); err == nil {
		links, err = workspace.ParseLinks(b)
		if err != nil {
			return nil, err
		}
	}
	ov, err := workspace.ComputeOverlay(root, reg, links)
	if err != nil {
		return nil, err
	}
	ov.GeneratedAt = nowRFC3339()
	if err := workspace.SaveOverlay(root, ov); err != nil {
		return nil, err
	}
	return ov, nil
}

// resolveWorkspaceRoot returns the workspace root: an explicit --root if present,
// else discovered by walking up from cwd. The returned args have --root removed.
func resolveWorkspaceRoot(args []string, stderr io.Writer) (root string, rest []string, code int) {
	rest = make([]string, 0, len(args))
	for i := 0; i < len(args); i++ {
		if args[i] == "--root" {
			if i+1 >= len(args) {
				fmt.Fprintln(stderr, "graffiti: --root requires a directory")
				return "", nil, 2
			}
			root = args[i+1]
			i++
			continue
		}
		rest = append(rest, args[i])
	}
	if root == "" {
		cwd, _ := os.Getwd()
		root = workspace.FindRoot(cwd)
		if root == "" {
			fmt.Fprintln(stderr, "graffiti: no workspace found (run `graffiti link` first, or pass --root)")
			return "", nil, 1
		}
	}
	return root, rest, 0
}

func runWorkspace(args []string, stdout, stderr io.Writer) int {
	root, rest, code := resolveWorkspaceRoot(args, stderr)
	if code != 0 {
		return code
	}
	if len(rest) == 0 {
		fmt.Fprintln(stderr, "graffiti: workspace <add|rm|list|render>")
		return 2
	}
	reg, err := workspace.LoadRegistry(root)
	if err != nil {
		fmt.Fprintf(stderr, "graffiti: %v\n", err)
		return 1
	}
	switch rest[0] {
	case "list":
		for _, m := range reg.Members {
			fmt.Fprintf(stdout, "%s\t%s\n", m.Alias, m.Path)
		}
		return 0
	case "render":
		return renderWorkspaceHTML(root, reg, stdout, stderr)
	case "rm":
		if len(rest) < 2 {
			fmt.Fprintln(stderr, "graffiti: workspace rm <alias>")
			return 2
		}
		if !workspace.RemoveMember(reg, rest[1]) {
			fmt.Fprintf(stderr, "graffiti: no member %q\n", rest[1])
			return 1
		}
	case "add":
		// graffiti workspace add <path> --as <alias>
		var path, alias string
		for i := 1; i < len(rest); i++ {
			if rest[i] == "--as" && i+1 < len(rest) {
				alias = rest[i+1]
				i++
			} else {
				path = rest[i]
			}
		}
		if path == "" {
			fmt.Fprintln(stderr, "graffiti: workspace add <path> [--as alias]")
			return 2
		}
		absPath, _ := filepath.Abs(path)
		if alias == "" {
			alias = aliasFor(absPath)
		}
		if _, err := os.Stat(filepath.Join(absPath, ".graffiti", "map.json")); err != nil {
			if _, berr := app.Build(absPath, nowRFC3339()); berr != nil {
				fmt.Fprintf(stderr, "graffiti: build %s: %v\n", absPath, berr)
				return 1
			}
		}
		rel, _ := filepath.Rel(root, absPath)
		h, _ := workspace.MapHash(absPath)
		workspace.AddMember(reg, workspace.Member{Alias: alias, Path: filepath.ToSlash(rel), MapHash: h})
	default:
		fmt.Fprintf(stderr, "graffiti: unknown workspace subcommand %q\n", rest[0])
		return 2
	}
	if err := workspace.SaveRegistry(root, reg); err != nil {
		fmt.Fprintf(stderr, "graffiti: %v\n", err)
		return 1
	}
	if _, err := computeAndSaveOverlay(root, reg); err != nil {
		fmt.Fprintf(stderr, "graffiti: %v\n", err)
		return 1
	}
	return 0
}

func runLinksCheck(args []string, stdout, stderr io.Writer) int {
	root, rest, code := resolveWorkspaceRoot(args, stderr)
	if code != 0 {
		return code
	}
	if len(rest) == 0 || rest[0] != "check" {
		fmt.Fprintln(stderr, "graffiti: links check")
		return 2
	}
	reg, err := workspace.LoadRegistry(root)
	if err != nil {
		fmt.Fprintf(stderr, "graffiti: %v\n", err)
		return 1
	}
	ov, err := computeAndSaveOverlay(root, reg)
	if err != nil {
		fmt.Fprintf(stderr, "graffiti: %v\n", err)
		return 1
	}
	fmt.Fprintf(stdout, "%d links OK.\n", len(ov.Links))
	if len(ov.Unresolved) > 0 {
		for _, l := range ov.Unresolved {
			fmt.Fprintf(stdout, "UNRESOLVED: %s -> %s\n", l.From, l.To)
		}
		return 1
	}
	return 0
}

// renderWorkspaceHTML writes the federated force-graph to
// <root>/.graffiti-workspace/workspace.html (projects as the tree's top level,
// cross-project links drawn). A missing overlay just renders members unlinked.
func renderWorkspaceHTML(root string, reg *workspace.Registry, stdout, stderr io.Writer) int {
	ov, _ := workspace.LoadOverlay(root) // nil overlay → members without cross-edges
	doc, err := workspace.CombinedDocument(root, reg, ov)
	if err != nil {
		fmt.Fprintf(stderr, "graffiti: %v\n", err)
		return 1
	}
	out := filepath.Join(root, workspace.WorkspaceDir, "workspace.html")
	if err := render.WriteWorkspaceHTML(doc, nowRFC3339(), out); err != nil {
		fmt.Fprintf(stderr, "graffiti: %v\n", err)
		return 1
	}
	fmt.Fprintf(stdout, "✓ wrote %s (%d nodes, %d edges, %d projects).\n", out, len(doc.Nodes), len(doc.Edges), len(reg.Members))
	return 0
}

func runFederateExplain(args []string, stdout, stderr io.Writer) int {
	root, _, code := resolveWorkspaceRoot(args, stderr)
	if code != 0 {
		return code
	}
	ov, err := workspace.LoadOverlay(root)
	if err != nil {
		fmt.Fprintf(stderr, "graffiti: %v\n", err)
		return 1
	}
	for _, l := range ov.Links {
		fmt.Fprintf(stdout, "%s -%s-> %s (%s, via %s)\n", l.From, l.Relation, l.To, l.Confidence, l.Via)
	}
	for _, l := range ov.Ambiguous {
		fmt.Fprintf(stdout, "AMBIGUOUS: %s -> %s (via %s)\n", l.From, l.To, l.Via)
	}
	return 0
}

func hasFlag(args []string, flag string) bool {
	for _, a := range args {
		if a == flag {
			return true
		}
	}
	return false
}

func stripFlag(args []string, flag string) []string {
	out := make([]string, 0, len(args))
	for _, a := range args {
		if a != flag {
			out = append(out, a)
		}
	}
	return out
}

// runQueryWorkspace loads the workspace, builds the combined alias-prefixed index,
// runs the LLM-free query over it, and appends a staleness nudge if any member
// changed since the overlay was computed. Args (after --workspace removed):
// optional --root <dir>, then the question (and optional [name], ignored in v1).
func runQueryWorkspace(args []string, stdout, stderr io.Writer) int {
	root, rest, code := resolveWorkspaceRoot(args, stderr)
	if code != 0 {
		return code
	}
	// the last non-flag arg is the question
	if len(rest) == 0 {
		fmt.Fprintln(stderr, "graffiti: query --workspace requires a question")
		return 2
	}
	question := rest[len(rest)-1]

	reg, err := workspace.LoadRegistry(root)
	if err != nil {
		fmt.Fprintf(stderr, "graffiti: %v\n", err)
		return 1
	}
	ov, err := workspace.LoadOverlay(root)
	if err != nil {
		fmt.Fprintf(stderr, "graffiti: %v\n", err)
		return 1
	}
	idx, err := workspace.CombinedIndex(root, reg, ov)
	if err != nil {
		fmt.Fprintf(stderr, "graffiti: %v\n", err)
		return 1
	}
	fmt.Fprint(stdout, query.Query(idx, question, query.DefaultTokenBudget))

	if stale, err := workspace.StaleMembers(root, reg, ov); err == nil && len(stale) > 0 {
		fmt.Fprintf(stdout, "\n(overlay stale: %s changed — run: graffiti update --workspace)\n", strings.Join(stale, ", "))
	}
	return 0
}

func serveWorkspace(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	root, _, code := resolveWorkspaceRoot(args, stderr)
	if code != 0 {
		return code
	}
	reg, err := workspace.LoadRegistry(root)
	if err != nil {
		fmt.Fprintf(stderr, "graffiti: %v\n", err)
		return 1
	}
	ov, err := workspace.LoadOverlay(root)
	if err != nil {
		fmt.Fprintf(stderr, "graffiti: %v\n", err)
		return 1
	}
	idx, err := workspace.CombinedIndex(root, reg, ov)
	if err != nil {
		fmt.Fprintf(stderr, "graffiti: %v\n", err)
		return 1
	}
	if err := mcp.NewServer(idx).Serve(stdin, stdout); err != nil {
		fmt.Fprintf(stderr, "graffiti: serve: %v\n", err)
		return 1
	}
	return 0
}

// updateWorkspace rebuilds members whose source changed since the registry's
// recorded hash, then recomputes the overlay. --links-only skips member rebuild.
func updateWorkspace(args []string, stdout, stderr io.Writer) int {
	linksOnly := hasFlag(args, "--links-only")
	root, _, code := resolveWorkspaceRoot(stripFlag(args, "--links-only"), stderr)
	if code != 0 {
		return code
	}
	reg, err := workspace.LoadRegistry(root)
	if err != nil {
		fmt.Fprintf(stderr, "graffiti: %v\n", err)
		return 1
	}
	rebuilt := 0
	if !linksOnly {
		for i := range reg.Members {
			dir := filepath.Join(root, filepath.FromSlash(reg.Members[i].Path))
			cur, err := workspace.MapHash(dir)
			if err != nil || cur != reg.Members[i].MapHash {
				if _, berr := app.Build(dir, nowRFC3339()); berr != nil {
					fmt.Fprintf(stderr, "graffiti: rebuild %s: %v\n", reg.Members[i].Alias, berr)
					return 1
				}
				if h, herr := workspace.MapHash(dir); herr == nil {
					reg.Members[i].MapHash = h
				}
				rebuilt++
			}
		}
		reg.GeneratedAt = nowRFC3339()
		if err := workspace.SaveRegistry(root, reg); err != nil {
			fmt.Fprintf(stderr, "graffiti: %v\n", err)
			return 1
		}
	}
	ov, err := computeAndSaveOverlay(root, reg)
	if err != nil {
		fmt.Fprintf(stderr, "graffiti: %v\n", err)
		return 1
	}
	fmt.Fprintf(stdout, "✓ Updated workspace: %d members rebuilt, overlay recomputed (%d links).\n", rebuilt, len(ov.Links))
	return 0
}

func runBuild(root string, stdout, stderr io.Writer) int {
	generatedAt := nowRFC3339()
	stats, err := app.Build(root, generatedAt)
	if err != nil {
		fmt.Fprintf(stderr, "graffiti: build failed: %v\n", err)
		return 1
	}
	fmt.Fprintf(stdout, "✓ Done. 0 API calls, $0.  %d files → %d nodes, %d edges, %d communities.\n",
		stats.Files, stats.Nodes, stats.Edges, stats.Communities)
	fmt.Fprintln(stdout, "  The 3 most interesting questions your map can answer:")
	for i, q := range stats.Questions {
		fmt.Fprintf(stdout, "    %d) %s\n", i+1, q)
	}
	return 0
}

// loadIndex loads <root>/.graffiti/map.json and builds the read-side Index.
func loadIndex(root string) (*store.Index, error) {
	path := mapPath(root)
	doc, err := store.Load(path)
	if err != nil {
		return nil, err
	}
	return store.NewIndex(doc), nil
}

func mapPath(root string) string {
	return filepath.Join(root, ".graffiti", "map.json")
}

func runQuery(root, question string, stdout, stderr io.Writer) int {
	idx, err := loadIndex(root)
	if err != nil {
		fmt.Fprintf(stderr, "graffiti: %v (run `graffiti build %s` first)\n", err, root)
		return 1
	}
	fmt.Fprint(stdout, query.Query(idx, question, query.DefaultTokenBudget))
	return 0
}

// serve runs the MCP stdio server. r/w/errW are injectable for tests; main wires
// os.Stdin/os.Stdout. Returns the exit code.
func serve(root string, r io.Reader, w, errW io.Writer) int {
	idx, err := loadIndex(root)
	if err != nil {
		fmt.Fprintf(errW, "graffiti: %v (run `graffiti build %s` first)\n", err, root)
		return 1
	}
	if err := mcp.NewServer(idx).Serve(r, w); err != nil {
		fmt.Fprintf(errW, "graffiti: serve: %v\n", err)
		return 1
	}
	return 0
}

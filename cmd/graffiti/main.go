// Command graffiti turns a code repository into a queryable directed knowledge graph.
package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/evgeniy-achin/graffiti/internal/app"
	"github.com/evgeniy-achin/graffiti/internal/mcp"
	"github.com/evgeniy-achin/graffiti/internal/query"
	"github.com/evgeniy-achin/graffiti/internal/store"
)

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
		// update is currently a full rebuild; the incremental AST-only rebuild
		// (spec §11) is a later optimization. Behaves exactly like `build`.
		root := "."
		if len(args) >= 3 {
			root = args[2]
		}
		return runBuild(root, stdout, stderr)
	case "query":
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
		root := "."
		if len(args) >= 3 {
			root = args[2]
		}
		return serve(root, stdin, stdout, stderr)
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
}

func runBuild(root string, stdout, stderr io.Writer) int {
	generatedAt := time.Now().UTC().Format(time.RFC3339)
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

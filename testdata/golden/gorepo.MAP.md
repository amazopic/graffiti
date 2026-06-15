# 001 — Map

_11 nodes, 13 edges, 3 communities. 0 API calls, $0._

## Start here

1. What does "greet/greet.go" do, and why is it touched by 3 things?
2. Why does "Formatter.Format" connect to "Hello" across subsystems?
3. How does the "Greet" subsystem (4 things) fit together?

## Landmarks (god nodes)

- **greet/greet.go** — touched by 3 things — change carefully.
- **Hello** — touched by 3 things — change carefully.
- **upper** — touched by 3 things — change carefully.
- **main** — touched by 3 things — change carefully.
- **main.go** — touched by 3 things — change carefully.
- **Formatter.Format** — touched by 2 things — change carefully.
- **greet/greet_helper.go** — touched by 2 things — change carefully.

## Districts

### Greet (4 things)

- `greet/greet.go` (file, greet/greet.go:1)
- `Hello` (function, greet/greet.go:5)
- `upper` (function, greet/greet.go:9)
- `strings` (module, greet/greet.go:3)

### Greet (3 things)

- `Formatter` (class, greet/greet_helper.go:3)
- `Formatter.Format` (method, greet/greet_helper.go:7)
- `greet/greet_helper.go` (file, greet/greet_helper.go:1)

### main (4 things)

- `main` (function, main.go:9)
- `main.go` (file, main.go:1)
- `greet` (module, main.go:6)
- `fmt` (module, main.go:4)

## Surprising connections

- `Formatter.Format` → `Hello` (calls, INFERRED)

## Confidence legend

- **EXTRACTED** — definite (verified from imports/syntax).
- **INFERRED** — inferred (same-package name match).
- **AMBIGUOUS** — guessed — verify.

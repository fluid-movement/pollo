---
title: Overview and Project Structure
description: >
  What pollo is, its design principles, how the codebase is organised,
  package responsibilities, and what the tool explicitly does not do.
topics:
  - tool purpose
  - design principles
  - project structure
  - package layout
  - module path
  - dependencies
  - non-goals
relevance: >
  Read this first. Essential context for any work on the project — understanding
  the design principles (stateless, JSON-only, atomic writes) prevents
  accidentally violating them. Also read when you need to know which package
  owns a concern, or to confirm something is out of scope.
---

# pollo — Overview

> A CLI tool for LLM-assisted translation of GNU gettext PO files.
> Designed to be used exclusively by an LLM agent (e.g. Claude Code).
> Never requires the LLM to load the full PO file into context.

`pollo` exposes a small set of deterministic subcommands that handle all PO
file I/O. The LLM's only job is to call `get`, produce a translation, and call
`set`. All parsing, validation, and file writing is handled by the tool.

The tool operates on a single `.po` or `.pot` file at a time. The file is the
source of truth — there is no external state or database.

## Design Principles

- All output is JSON on stdout, always. No mixed output.
- Structured errors (user errors, validation errors) go to stdout as JSON with
  `"ok": false`. Only unexpected internal failures go to stderr.
- Every command is stateless and idempotent — safe to retry.
- The file is never left in a corrupt state (atomic writes via temp file +
  rename).
- The tool never writes anything to the PO file except confirmed translations.
  Skip state, session state, and run metadata are never persisted to the PO
  file.

---

## Project Structure

```
pollo/
├── main.go
├── go.mod
├── go.sum
├── cmd/                  // CLI layer: flag parsing, JSON output, no domain logic
│   ├── root.go
│   ├── stats.go
│   ├── get.go
│   └── set.go
└── po/                   // Domain layer: all PO parsing, writing, and validation
    ├── types.go
    ├── parser.go
    ├── writer.go
    ├── wrap.go
    ├── validate.go
    ├── plural.go
    ├── skipfile.go
    ├── po_test.go
    └── testdata/
        ├── fixture_a.po
        ├── fixture_b.po
        ├── fixture_c.po
        ├── fixture_d.po
        └── fixture_e.po
```

### Module and Go Version

- Module path: `github.com/fluid-movement/pollo`
- Go version: `1.26` (current stable release)
- Single third-party dependency: `github.com/spf13/cobra` for CLI
- The PO parser is hand-written; no third-party PO library

Run `go mod tidy` to fetch and pin dependencies. No other setup required.

### Package Responsibilities

- `main` — entry point only; delegates immediately to `cmd`
- `cmd` — flag parsing, calling `po` functions, serialising JSON responses;
  contains no domain logic
- `po` — all PO domain logic: parsing, writing, validation, plural rules, skip
  file I/O; no filesystem access beyond what is passed in via interfaces

The `po` package must not import `cmd`. Dependency flows one way:
`main` → `cmd` → `po`.

---

## Non-Goals

- No `.mo` binary file support
- No `msgmerge` / `.pot` template merging
- No translation memory or glossary
- No batch import/export (XLIFF, CSV, etc.)
- No multi-file operation in a single invocation
- No authentication, no network access
- No multi-entry `get` responses (one get → one translate → one set)
- No persistent state written to the PO file by the tool

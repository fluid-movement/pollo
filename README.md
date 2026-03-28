# pollo

A CLI tool for LLM-assisted translation of GNU gettext PO files.

Designed to be driven by an LLM agent (e.g. Claude Code). The agent calls `get` to fetch the next untranslated string, produces a translation, then calls `set` to write it back — without ever loading the full PO file into context.

## How it works

```
pollo stats locales/de.po
pollo get   locales/de.po
pollo set   locales/de.po --id "Save changes" --translation "Änderungen speichern"
```

All output is JSON on stdout. The file is never left in a corrupt state (atomic writes via temp file + rename).

## Commands

| Command | Description |
|---|---|
| `pollo stats <file>` | Translation progress: total, translated, fuzzy, untranslated |
| `pollo get <file>` | Next entry needing translation (fuzzy-first by default) |
| `pollo set <file>` | Write a confirmed translation; clears the fuzzy flag |

`get` supports `--order`, `--skip-ids-file`, `--include-translated`, and lookup by `--id`. `set` validates placeholders and supports plural forms via `--translations`.

## Design

- Output is always a single JSON object — nothing else on stdout, ever
- Stateless and idempotent — every command is safe to retry
- The PO file is the only source of truth — no external state or database
- Skip state is agent-managed (a plain text file); the PO file is never modified for it

## Install

```bash
go install github.com/fluid-movement/pollo@latest
```

Or build from source:

```bash
git clone https://github.com/fluid-movement/pollo
cd pollo
go build -o pollo .
```

Requires Go 1.26

## Status

Under active development. See [SPEC.md](SPEC.md) for the full technical specification.

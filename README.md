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

### Download a prebuilt binary

Download the latest release for your platform from the
[Releases page](https://github.com/fluid-movement/pollo/releases).

Extract and place the `pollo` binary somewhere on your `$PATH`:

```bash
# macOS arm64
tar -xzf pollo_darwin_arm64.tar.gz
mv pollo /usr/local/bin/
```

### Via go install (requires Go 1.26)

```bash
go install github.com/fluid-movement/pollo@latest
```

### Build from source

```bash
git clone https://github.com/fluid-movement/pollo
cd pollo
go build -o pollo .
```

## Using with Claude Code

pollo is designed to be driven by a Claude Code agent. `SKILL.md` is bundled
in every release archive — it tells Claude exactly how to run the translation
loop.

To set it up:

```bash
# After extracting the release archive:
cp SKILL.md /path/to/your/project/.claude/skills/pollo.md
```

Then ask Claude to translate a PO file and it will follow the
get → translate → set loop automatically.

## Status

Under active development. See [docs/](docs/README.md) for the full technical specification.

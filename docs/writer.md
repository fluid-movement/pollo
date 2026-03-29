---
title: Writer Behaviour and String Conventions
description: >
  How the PO writer serialises entries: field ordering, string escaping rules,
  line wrapping to 76 columns, atomic file writes, and the string identity
  convention (raw unescaped values in flags, file-based value flags, UTF-8
  validation).
topics:
  - writer
  - PO serialisation
  - field ordering
  - string escaping
  - line wrapping
  - 76-column limit
  - atomic write
  - temp file rename
  - string identity
  - raw unescaped values
  - file-based flags
  - UTF-8 validation
relevance: >
  Read when modifying writer.go or wrap.go, debugging unexpected output format,
  understanding how a `set` write mutates an entry, or when passing string
  values to CLI flags and need to know what form they should take.
---

# Writer Behaviour and String Conventions

## Strategy

The writer reconstructs each entry from its decoded fields (AST reconstruction).
It does not attempt byte-for-byte preservation. Output must pass `msgfmt -c`.

## Field Order Per Entry

Each entry is serialised in this order, omitting fields that are empty or nil:

```
[blank line before each entry except the first]
# <translator comment>            one line per embedded newline
#. <extracted comment>            one line per embedded newline
#: <references>                   all references on one line, space-separated
#, <flags>                        all flags on one line, comma-separated
#| msgid "<previous msgid>"       only if previous msgid is set
#| msgid_plural "<prev plural>"   only if previous msgid_plural is set
msgctxt "<msgctxt>"               only if msgctxt is set
msgid "<msgid>"
msgid_plural "<msgid_plural>"     only if plural entry
msgstr "<msgstr>"                 singular only
msgstr[0] "<form0>"               plural only
msgstr[1] "<form1>"               plural only
...
```

The header entry is serialised first with no preceding blank line. Obsolete
nodes are written verbatim at their original position in the node list.

The file ends with a single trailing newline.

## String Escaping

When encoding a decoded string to a PO quoted value, apply in order:

1. `\` → `\\`
2. `"` → `\"`
3. newline → `\n`
4. tab → `\t`
5. carriage return → `\r`
6. BEL (0x07) → `\a`
7. BS (0x08) → `\b`
8. FF (0x0C) → `\f`
9. VT (0x0B) → `\v`

No other characters are escaped. Non-ASCII Unicode is written as-is (UTF-8).

## Line Wrapping

Serialised strings are wrapped to 76 columns (rune count, not bytes) to match
GNU gettext default behaviour and produce clean diffs after `msgmerge` re-runs.

**Rules:**

- If the escaped value contains no `\n` sequences and fits within 76 runes on
  a single line (including keyword and surrounding quotes), write as a single
  line.
- Otherwise, write `keyword ""` on the first line, then the value as one or
  more continuation lines in `"..."` form.
- Split primarily on `\n` escape sequences (keeping `\n` at the end of each
  segment), then further split any segment exceeding 76 runes at the last space
  boundary before the limit.
- If a segment has no space boundary within the limit (e.g. a URL), emit the
  oversized line without splitting — matching GNU gettext behaviour.
- An empty value always writes as a single line `keyword ""`.

## On `set` Write

When a confirmed translation is written:
- Replace `msgstr` / `msgstr[n]` with the provided values.
- Remove `fuzzy` from the entry's flags. If flags become empty, the `#,` line
  is omitted.
- Clear previous msgid and previous msgid_plural. The `#|` lines are removed.

## Atomic Write

The full reconstructed file is written to a temp file in the same directory as
the target, then renamed over it atomically (POSIX `rename`). On any failure
the temp file is removed and the original is untouched.

---

## String Identity and Flag Value Convention

### Raw Unescaped Values

All string flag values (`--id`, `--translation`, `--context`) accept the raw
unescaped string — the actual text content, not the PO-escaped representation.

Example: `msgid "Save\nchanges"` in the file stores the actual string
`Save` + newline + `changes`. The agent must pass that real newline to `--id`.

### File-Based Value Flags

Every string value flag has a `*-file` variant that reads the value from a
file path (raw bytes, no unescaping). This eliminates shell-escaping concerns
for all string content.

| Value flag | File variant |
|---|---|
| `--id` | `--id-file` |
| `--translation` | `--translation-file` |
| `--translations` | `--translations-file` |
| `--context` | `--context-file` |

Each pair is mutually exclusive. For `--translations-file`, the file must
contain a valid JSON array of strings.

### UTF-8 Validation

All string values from flags or flag-files are validated as UTF-8. Invalid
bytes produce `{"ok": false, "error": "..."}` and exit code 1.

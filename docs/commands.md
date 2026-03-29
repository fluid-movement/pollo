---
title: Commands, Output Schemas, and Placeholder Validation
description: >
  Full specification of the three CLI commands (stats, get, set): their flags,
  output JSON schemas with examples, exit codes, placeholder validation rules
  for format strings, and the output format contract (stdout-only JSON,
  snake_case keys, null vs omitted fields).
topics:
  - stats command
  - get command
  - set command
  - CLI flags
  - JSON output schema
  - exit codes
  - done sentinel
  - fuzzy entry
  - plural entry
  - skip mechanism
  - placeholder validation
  - c-format
  - python-format
  - python-brace-format
  - sh-format
  - output format contract
  - snake_case
  - null fields
relevance: >
  The primary reference for anyone calling or implementing the commands.
  Read when you need the exact JSON shape of a response, want to understand
  flag semantics, or are implementing/debugging placeholder validation.
---

# Commands

## `pollo stats <file>`

Report translation progress for the file.

**Output:**
```json
{
  "file": "locales/de.po",
  "language": "de",
  "language_name": "German",
  "total": 142,
  "translated": 98,
  "fuzzy": 7,
  "untranslated": 37,
  "remaining": 44,
  "parse_warnings": []
}
```

- `remaining` = `fuzzy` + `untranslated`
- `total` = `translated` + `fuzzy` + `untranslated`
- `parse_warnings` is always an array; empty if no warnings

**Exit codes:** `0` success, `1` fatal error.

---

## `pollo get <file> [flags]`

Return the next entry needing translation, or a specific entry by ID.

**Flags:**

| Flag | Description |
|---|---|
| `--id` | Fetch specific entry by unescaped msgid |
| `--id-file` | Read msgid from file |
| `--context` | msgctxt for the target entry |
| `--context-file` | Read msgctxt from file |
| `--skip-ids-file` | Path to skip file (see [agent-usage.md](agent-usage.md)); missing file = empty list |
| `--order` | `fuzzy-first` (default) or `untranslated-first` |
| `--include-translated` | Return all entries regardless of state |

`--id`/`--id-file` are mutually exclusive with `--order`, `--skip-ids-file`,
and `--include-translated`.

**Iteration behaviour (no `--id`):**

With `--order fuzzy-first` (default): all fuzzy entries in file order, then all
untranslated entries in file order. Skip entries whose key appears in the skip
file.

With `--order untranslated-first`: untranslated before fuzzy.

With `--include-translated`: all non-header, non-obsolete entries in file order,
still honouring the skip file.

When no qualifying entry remains, return the done sentinel.

**Output — singular entry:**
```json
{
  "done": false,
  "file": "locales/de.po",
  "language": "de",
  "language_name": "German",
  "msgid": "Save changes",
  "msgid_plural": null,
  "msgctxt": null,
  "state": "untranslated",
  "translator_comment": null,
  "extracted_comment": "Button label in the save dialog",
  "flags": [],
  "previous_msgid": null,
  "previous_msgid_plural": null,
  "current_msgstr": "",
  "current_msgstr_plural": null,
  "plural_count": null,
  "remaining": 44,
  "remaining_fuzzy": 7,
  "remaining_untranslated": 37,
  "parse_warnings": []
}
```

**Output — plural entry:**
```json
{
  "done": false,
  "file": "locales/de.po",
  "language": "de",
  "language_name": "German",
  "msgid": "%d item",
  "msgid_plural": "%d items",
  "msgctxt": null,
  "state": "untranslated",
  "translator_comment": null,
  "extracted_comment": null,
  "flags": ["c-format"],
  "previous_msgid": null,
  "previous_msgid_plural": null,
  "current_msgstr": null,
  "current_msgstr_plural": ["", ""],
  "plural_count": 2,
  "remaining": 44,
  "remaining_fuzzy": 7,
  "remaining_untranslated": 36,
  "parse_warnings": []
}
```

**Output — fuzzy entry:**
```json
{
  "done": false,
  "file": "locales/de.po",
  "language": "de",
  "language_name": "German",
  "msgid": "Delete selected items",
  "msgid_plural": null,
  "msgctxt": null,
  "state": "fuzzy",
  "translator_comment": null,
  "extracted_comment": null,
  "flags": [],
  "previous_msgid": "Delete item",
  "previous_msgid_plural": null,
  "current_msgstr": "Element löschen",
  "current_msgstr_plural": null,
  "plural_count": null,
  "remaining": 7,
  "remaining_fuzzy": 7,
  "remaining_untranslated": 0,
  "parse_warnings": []
}
```

Note: `flags` never contains `"fuzzy"` — state is already conveyed by
`"state": "fuzzy"`.

**Output — done sentinel:**
```json
{
  "done": true,
  "file": "locales/de.po",
  "language": "de",
  "language_name": "German",
  "remaining": 0,
  "remaining_fuzzy": 0,
  "remaining_untranslated": 0
}
```

**Schema rules:**
- Every non-done response includes both `current_msgstr` and
  `current_msgstr_plural`; one will be `null`.
- All optional fields are always present as `null`, never absent.
- `parse_warnings` is always an array.

**Exit codes:** `0` success (including done), `1` fatal error,
`2` entry not found when `--id`/`--id-file` was given.

---

## `pollo set <file> [flags]`

Write a confirmed translation; clear the fuzzy flag.

**Flags:**

| Flag | Required | Description |
|---|---|---|
| `--id` | yes* | Unescaped msgid of the target entry |
| `--id-file` | yes* | Read msgid from file |
| `--context` | no† | msgctxt of the target entry |
| `--context-file` | no† | Read msgctxt from file |
| `--translation` | yes‡ | Translated string (singular) |
| `--translation-file` | yes‡ | Read translation from file |
| `--translations` | yes‡ | JSON array of plural forms |
| `--translations-file` | yes‡ | Read JSON array from file |

\* Exactly one of `--id` or `--id-file`.
† Exactly one of `--context` or `--context-file`, only if entry has msgctxt.
‡ Exactly one of the four translation flags.

**Hard errors (`ok: false`, exit 1):**

1. Entry not found.
2. Singular entry + plural translation flag.
3. Plural entry + singular translation flag.
4. Plural array length ≠ `nplurals` for the file.
5. Any string value is not valid UTF-8.

**Soft warnings (`ok: true` + `"warning"` field):**

6. Placeholder mismatch when a known format flag is present (see Placeholder Validation below).

An empty string `""` is a valid translation value.

**Remaining counts** in the response are computed from in-memory state *after*
applying the write. The just-translated entry is no longer counted.

**Output — success:**
```json
{
  "ok": true,
  "file": "locales/de.po",
  "msgid": "Save changes",
  "remaining": 43,
  "remaining_fuzzy": 7,
  "remaining_untranslated": 36,
  "parse_warnings": []
}
```

**Output — success with warning:**
```json
{
  "ok": true,
  "file": "locales/de.po",
  "msgid": "%d item",
  "warning": "placeholder mismatch: source has [%d] but translation has []",
  "remaining": 43,
  "remaining_fuzzy": 7,
  "remaining_untranslated": 36,
  "parse_warnings": []
}
```

**Output — error:**
```json
{
  "ok": false,
  "error": "entry not found: \"Save changes\""
}
```

**Exit codes:** `0` success (with or without warning), `1` any error.

---

## Skip Mechanism

There is no `skip` subcommand. Skip state is managed by the agent, not the
tool. Writing non-standard flags to the PO file would break `msgmerge`
pipelines and produce spurious diffs.

The agent maintains a plain text skip file (see [agent-usage.md](agent-usage.md))
and passes it to `get` via `--skip-ids-file`. The PO file is never modified for
skip purposes.

---

## Placeholder Validation

Validation runs on `set` when the entry's flags include a known format marker.
The comparison is multiset equality: source and translation must contain the
same placeholders the same number of times (order does not matter).

`%%` is excluded from both sides — it is a literal percent sign, not a
placeholder.

For plural entries: `msgstr[0]` is validated against `msgid`; all other forms
are validated against `msgid_plural`.

### Supported Flags and Patterns

| Flag | Pattern matches | Example |
|---|---|---|
| `c-format` | Full POSIX printf: `%d`, `%s`, `%ld`, `%02.4f`, `%1$s` (positional), etc. | `"Hello, %s! You have %d messages"` |
| `objc-format` | Same as `c-format` | |
| `python-format` | printf-style plus named groups: `%(name)s` | `"Hello, %(user)s!"` |
| `python-brace-format` | Brace placeholders: `{0}`, `{name}`, `{value:.2f}` | `"Hello, {name}!"` |
| `php-format` | printf-style as used by PHP | `"You have %d items"` |
| `javascript-format` | Same as `c-format` | |
| `sh-format` | Shell variable references: `$VAR`, `${VAR}` | `"Hello, $USER!"` |

For unrecognised format flags, skip validation silently.

### Warning Format

```
placeholder mismatch: source has [%d, %s] but translation has [%s]
```

Placeholders in each bracket are listed in sorted order for determinism.

---

## Output Format Contract

1. **stdout** contains exactly one JSON object per command invocation,
   terminated by a newline. Nothing else is ever written to stdout.
2. **stderr** is reserved for unexpected internal failures only. All user
   errors and validation errors go to stdout as `{"ok": false, ...}`.
3. No ANSI codes, no colour, no progress spinners, no interactive prompts, no
   log lines on either stream.
4. All JSON keys use `snake_case`.
5. Optional fields are always present as `null`, never omitted. The agent
   relies on a fixed schema shape per command.
6. `parse_warnings` is always a JSON array (`[]` when empty), never `null`.
7. The `warning` field (placeholder mismatch) is present only when raised;
   otherwise omitted entirely. This lets the agent use a simple presence check.
8. HTML-special characters (`<`, `>`, `&`) in string values must not be
   escaped to Unicode escape sequences in JSON output. They must appear as
   literal characters.

---
title: LLM Agent Usage Pattern
description: >
  How an LLM agent should use pollo to translate a PO file: the stats → loop
  (get → translate → set) workflow, how to handle fuzzy and plural entries,
  the skip file format for entries that cannot be translated, shell escaping
  recommendations, and worked example invocations.
topics:
  - agent workflow
  - translation loop
  - fuzzy entry handling
  - plural entry handling
  - skip file
  - skip file format
  - entry key construction
  - shell escaping
  - process substitution
  - example invocations
relevance: >
  The primary reference for an LLM agent driving a translation session.
  Read this before starting any translation task. Also read when you need
  the skip file format or guidance on shell-safe flag usage.
---

# LLM Agent Usage Pattern

## Translation Workflow

```
1.  pollo stats <file>
    → Report to user: "X of Y strings translated (F fuzzy, U untranslated)"

2.  Prepare a skip file path (optional but recommended):
    SKIP_FILE="/tmp/pollo-skip-$(printf '%s' '<file>' | sha256sum | cut -c1-16).txt"
    touch "$SKIP_FILE"

3.  Loop:

    a.  pollo get <file> [--skip-ids-file "$SKIP_FILE"]
    b.  If response.done == true → break

    c.  Build the translation:
        - Target language:   response.language_name
          (use the human-readable name in your prompt, e.g. "Translate to German (Austria)")
        - Source string:     response.msgid
        - Use these context signals when non-null:
            response.translator_comment  — explicit note from the developer
            response.extracted_comment   — source code context (e.g. "button label")
            response.msgctxt             — UI location; disambiguates identical source strings
        - If response.flags contains a format marker (e.g. "c-format",
          "python-brace-format"): preserve every placeholder exactly as-is.
        - If response.state == "fuzzy":
            response.previous_msgid shows the old source string
            response.current_msgstr is the outdated translation
            → Update the translation to match response.msgid, using
              response.current_msgstr as a starting point.
        - If response.msgid_plural != null:
            → Produce exactly response.plural_count translation forms.
            → Pass as a JSON array to --translations / --translations-file.

    d.  Write the translation (use file flags for safety):
        pollo set <file> \
          --id-file <(printf '%s' "$MSGID") \
          --translation-file <(printf '%s' "$TRANSLATION")
        — or for plurals —
        pollo set <file> \
          --id-file <(printf '%s' "$MSGID") \
          --translations-file <(printf '%s' "$JSON_ARRAY")

    e.  If set response.ok == false → report error to user, stop.
    f.  If set response.warning != null → log the warning, continue.

    g.  If you cannot translate this entry:
        printf '%s\n' "$ENTRY_KEY" >> "$SKIP_FILE"
        (continue loop without calling set)

4.  pollo stats <file>
    → Report final result to user
```

## Constructing ENTRY_KEY for the Skip File

```bash
# No msgctxt:
ENTRY_KEY="$MSGID"

# With msgctxt (U+0004 separator):
ENTRY_KEY="${MSGCTXT}"$'\x04'"${MSGID}"
```

## Shell Escaping Recommendation

Always prefer `--id-file` and `--translation-file` for programmatic use.
Process substitution is the cleanest approach:

```bash
pollo set locales/de.po \
  --id-file <(printf '%s' "$MSGID") \
  --translation-file <(printf '%s' "$TRANSLATION")
```

For shells without process substitution, write to temp files instead.

---

## Example Invocations

```bash
# Check progress
pollo stats locales/de.po

# Get next entry needing work (fuzzy first by default)
pollo get locales/de.po

# Get next untranslated entry first
pollo get locales/de.po --order untranslated-first

# Get all entries including already-translated ones
pollo get locales/de.po --include-translated

# Fetch a specific entry
pollo get locales/de.po --id "Save changes"

# Fetch an entry whose msgid contains a literal newline
pollo get locales/de.po --id-file <(printf 'Line one\nLine two')

# Translate a singular string
pollo set locales/de.po \
  --id-file <(printf '%s' "Save changes") \
  --translation-file <(printf '%s' "Änderungen speichern")

# Translate a string with msgctxt
pollo set locales/de.po \
  --id "Back" \
  --context "navigation button" \
  --translation "Zurück"

# Translate a plural string (German: 2 forms)
pollo set locales/de.po \
  --id "%d item" \
  --translations '["Ein Element", "%d Elemente"]'

# Skip an untranslatable entry (agent-managed; PO file unchanged)
printf '%s\n' "INTERNAL_KEY" >> /tmp/skipped.txt
pollo get locales/de.po --skip-ids-file /tmp/skipped.txt
```

---

## Skip File Format

The skip file is UTF-8 plain text, one record per line (`\n`-terminated; `\r\n`
also accepted).

Each line is one of:
- `<key>` — entry key only
- `<key><TAB><reason>` — key with optional human-readable reason
- blank line or line starting with `#` — comment, ignored

`<key>` is the entry's canonical key (`Entry.Key()`) with these escape
sequences applied for safe line-based storage:

| Character in key | File representation |
|---|---|
| `\n` (newline) | `\n` (two characters: backslash + n) |
| `\r` (carriage return) | `\r` (two characters: backslash + r) |
| `\\` (backslash) | `\\` (two backslashes) |
| tab | literal tab — **not** escaped (used as field separator) |
| U+0004 (EOT, context separator) | literal `\x04` byte — preserved as-is |

**Behaviour:** if the file does not exist, treat as an empty skip list (no
error). The tool never writes or modifies the skip file.

**Example:**
```
# Skip file for locales/de.po
INTERNAL_CONFIG_KEY	Not user-visible; do not translate
Some string with\nnewline	msgid contains a newline
```

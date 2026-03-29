---
name: pollo
description: >
  Translate GNU gettext PO files using the pollo CLI. Use when the user asks
  to translate a PO file, fill in missing translations, review fuzzy entries,
  or check translation progress. Triggers: "translate this PO file", "fill in
  the missing translations", "review fuzzy strings", "how complete is the
  translation", "translate locales/X.po".
requires: pollo binary in $PATH
---

# pollo

Reads and writes PO files one entry at a time. You never load the full file.
All commands write JSON to stdout; ignore stderr.

## Loop

```
1. pollo stats <file>        → report progress to user
2. SKIP_FILE=$(mktemp)
3. loop:
   a. pollo get <file> --skip-ids-file "$SKIP_FILE"
   b. if response.done == true → break
   c. translate (see below)
   d. pollo set <file> --id-file <(printf '%s' "$MSGID") --translation-file <(printf '%s' "$TRANSLATION")
      plural: --translations-file <(printf '%s' "$JSON_ARRAY")
      if response.ok == false → report error, stop
      if response.warning present → note it, continue
   e. if you cannot translate:
      printf '%s\n' "$ENTRY_KEY" >> "$SKIP_FILE"   # no msgctxt
      printf '%s\n' "${MSGCTXT}"$'\x04'"${MSGID}" >> "$SKIP_FILE"  # with msgctxt
4. pollo stats <file>        → report final progress to user
```

**Always use `--id-file` / `--translation-file` with process substitution** —
direct flags corrupt values containing newlines, quotes, or backslashes.

## Translating an entry

Use every non-null field from the `get` response:

- **`language_name`** — target language (e.g. `"German (Austria)"`). If empty,
  ask the user before proceeding.
- **`msgid`** — source string to translate
- **`translator_comment`** — developer note written for you; highest priority
- **`extracted_comment`** — source code context (e.g. `"Button label"`)
- **`msgctxt`** — disambiguates identical source strings in different UI locations
- **`flags`** — if contains `c-format`, `python-brace-format`, etc.: preserve
  every placeholder (`%s`, `{name}`, `$VAR`) exactly as-is
- **`state == "fuzzy"`** — source changed; `previous_msgid` shows what it was,
  `current_msgstr` is the outdated translation. Update it to match the new msgid.
- **`msgid_plural != null`** — produce exactly `plural_count` forms as a JSON
  array. Example (German): `["Ein Element", "%d Elemente"]`
---
name: pollo
description: >
  Translate GNU gettext PO files using the pollo CLI. Use when the user asks
  to translate a PO file, fill in missing translations, review fuzzy entries,
  or check translation progress. Triggers: "translate this PO file", "fill in
  the missing translations", "review fuzzy strings", "how complete is the
  translation", "translate locales/X.po".
requires: pollo in $PATH — go install github.com/fluid-movement/pollo@latest
---

# pollo

Reads and writes PO files one entry at a time. You never load the full file.
All commands write JSON to stdout; ignore stderr.

## Loop

Each step is a separate tool call. Never combine into a compound assignment like
`SKIP_FILE=$(mktemp)` or `ENTRY=$(pollo get ...)` — commands starting with a
variable assignment bypass the permission allow-list and get denied.

```
1. pollo stats <file>
   → report progress to user

2. Run via Bash: mktemp
   → note the output path (call it SKIP_FILE in your head)

3. loop:
   a. Run via Bash: pollo get <file> --skip-ids-file <SKIP_FILE_PATH> > /tmp/pollo_entry.json
      → the JSON is now in /tmp/pollo_entry.json

   b. Run via Bash: python3 -c "import json; d=json.loads(open('/tmp/pollo_entry.json').read(), strict=False); print(d['done'])"
      → if output is "True" → break

   c. Translate (read msgid and other fields from /tmp/pollo_entry.json using Read tool)
      → decide TRANSLATION string

   d. Run via Bash: python3 -c "import json; d=json.loads(open('/tmp/pollo_entry.json').read(), strict=False); open('/tmp/pollo_msgid.txt','w').write(d['msgid'])"
      → writes exact msgid bytes (no trailing newline, no character transcription risk)

      Write TRANSLATION to /tmp/pollo_translation.txt using the Write tool

      For singular:
        Run via Bash: pollo set <file> --id-file /tmp/pollo_msgid.txt --translation-file /tmp/pollo_translation.txt

      For plural (msgid_plural != null):
        Write the JSON array of plural forms to /tmp/pollo_translations.json using the Write tool
        Run via Bash: pollo set <file> --id-file /tmp/pollo_msgid.txt --translations-file /tmp/pollo_translations.json

      if response.ok == false → report error, stop
      if response.warning present → note it, continue

   e. if you cannot translate this entry:
      Run via Bash: python3 -c "import json; d=json.loads(open('/tmp/pollo_entry.json').read(), strict=False); k=d.get('msgctxt') or ''; mid=d['msgid']; print((k+chr(4)+mid) if k else mid)" >> <SKIP_FILE_PATH>

4. pollo stats <file>
   → report final progress to user
```

**Use python3 to extract msgid** — never transcribe it manually into the Write
tool. Manual transcription risks replacing Unicode characters (e.g. curly
quotes `"`) with ASCII look-alikes (`"`), causing silent lookup failures.

**pollo set / pollo get trim trailing newlines** from `--id-file` and
`--translation-file` automatically, so the Write tool's trailing newline on
translation files is harmless.

**Note on JSON parsing:** pollo's JSON output may contain literal newline
characters inside string values. Always use `strict=False`.

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

---
title: Go Implementation Requirements and Testing
description: >
  Go-specific coding requirements (version, style, error handling, JSON
  encoding, testability constraints), the full required test coverage checklist,
  test fixture PO files (A–E) with their expected parse state, and build/install
  instructions.
topics:
  - Go requirements
  - Go version
  - code style
  - error handling
  - JSON encoding
  - testability
  - io.Reader / io.Writer
  - atomic writes
  - no global state
  - test coverage
  - test fixtures
  - fixture_a.po
  - fixture_b.po
  - fixture_c.po
  - fixture_d.po
  - fixture_e.po
  - build
  - install
  - go test
relevance: >
  Read when writing or reviewing Go code in this project, adding tests,
  or setting up the build. The test fixtures section is the authoritative
  source for what each testdata file contains and what state is expected
  after parsing.
---

# Go Implementation Requirements

- **Go version:** `1.26` (current stable release as of February 2026)
- **Style:** Follow the official Go style guide and *Effective Go*. Code must
  pass `go vet` and `staticcheck` with no warnings.
- **Formatting:** All code formatted with `gofmt`.
- **Error handling:** Use `fmt.Errorf` with `%w` for wrapping. Functions in
  the `po` package return errors; they never call `os.Exit` or `log.Fatal`.
  Only `cmd` handlers terminate the process.
- **Testability:** The `po` package must accept `io.Reader`/`io.Writer`
  interfaces rather than file paths, so all domain logic is testable without
  filesystem access. File I/O is the `cmd` layer's responsibility.
- **Atomicity:** File writes use a temp file + `os.Rename` in the same
  directory as the target.
- **JSON:** Use `encoding/json`. HTML escaping must be disabled on the encoder
  so that `<`, `>`, and `&` in translation strings are not mangled.
- **Modern stdlib:** Use `slices`, `maps`, and other stdlib packages introduced
  in recent Go versions where appropriate. Avoid reinventing what the standard
  library provides.
- **No global state:** No package-level variables that are mutated at runtime.
- **No external PO library:** The parser is hand-written as specified.

---

## Build and Install

```bash
# Fetch dependencies
go mod tidy

# Run all tests
go test ./...

# Build binary
go build -o pollo .

# Install to $GOPATH/bin
go install .
```

No code generation, no Makefile, no build tags required.

---

## Testing Requirements

All `po` package tests live in a single test file using the external test
package (`po_test`), testing only the exported API.

Test fixtures in `po/testdata/` are loaded by path (see fixture definitions
below).

### Required Test Coverage

**Parser:**
- Minimal file (header only)
- Single fully-translated entry
- Untranslated entry (empty msgstr)
- Plural entry with all forms
- Fuzzy entry with `#| msgid`
- Fuzzy plural entry with `#| msgid_plural`
- Entry with `msgctxt`
- Multi-line msgid and msgstr
- Multi-line `#| msgid` continuation
- `#~` blocks interleaved between active entries
- `#~` block at end of file
- Translator comment (`# `)
- Extracted comment (`#.`)
- Source references (`#:`)
- Multiple flags on one `#,` line
- Duplicate key → two entries retained + warning
- Non-UTF-8 charset header → fatal error
- Absent charset header → accepted
- Empty input → zero entries, no error
- Windows `\r\n` line endings → parsed identically to `\n`
- No header entry
- Non-contiguous `msgstr[n]` indices → fatal error
- Unrecognised escape sequence → pass-through + warning

**Entry state logic:**
- Singular translated
- Singular untranslated (empty)
- Singular untranslated (whitespace-only)
- Fuzzy with non-empty msgstr → state is "fuzzy" not "translated"
- Plural all forms filled → translated
- Plural one form empty → untranslated

**Writer:**
- Parse then write minimal file; output is valid PO
- Parse fixture C (full coverage fixture), write, re-parse; all fields round-trip correctly
- After `set`: fuzzy flag removed, `#|` lines absent, `#,` line omitted if no remaining flags
- Obsolete blocks appear at their original position (not moved to end)

**Line wrapping:**
- Short string → single-line output
- String exactly at column limit → single-line
- String one rune over limit → multi-line
- String containing `\n` sequences → multi-line always
- Long token with no space boundary → oversized line emitted without splitting
- CJK characters → counted as one rune each, not three bytes
- Empty value → `keyword ""\n`
- `col = 0` → wrapping disabled

**Placeholder validation:**
- Matching `c-format` placeholders → no warning
- Mismatched `c-format` → warning with correct diff
- Positional args `%1$s %2$d` handled correctly
- `%%` excluded from both sides
- Python brace-format match and mismatch
- Unrecognised format flag → no warning
- No format flag → no warning

**Plural count resolution:**
- From `Plural-Forms` header
- CLDR fallback for German (`de` → 2)
- CLDR fallback for Arabic (`ar` → 6)
- CLDR fallback for Japanese (`ja` → 1)
- Language absent + plural entries → fatal error
- Language absent + no plural entries → `nplurals = 1`, no error

**Skip file:**
- Empty file → empty result
- Simple msgid entries
- Key with `msgctxt` separator
- Key containing escaped newline
- Entry with tab-separated reason
- Missing file → empty result, no error

**Integration test:**
Simulate a full agent workflow using fixture B:
1. Parse and assert initial stats (1 translated, 1 fuzzy, 2 untranslated).
2. Get first entry (fuzzy-first order) — assert it is the fuzzy entry.
3. Apply a `set` — assert fuzzy flag and `#|` lines removed.
4. Continue until done. Write to buffer, re-parse, assert all entries translated.

---

## Test Fixtures

Store these files under `po/testdata/`. Tests reference them by path.

---

### Fixture A — Minimal (header only)

**File:** `po/testdata/fixture_a.po`

```po
msgid ""
msgstr ""
"Content-Type: text/plain; charset=UTF-8\n"
"Content-Transfer-Encoding: 8bit\n"
"Language: de\n"
"Plural-Forms: nplurals=2; plural=(n != 1);\n"
```

**Expected state after parse:**
- Header present; language = `"de"`; plural count = `2`
- Zero translatable entries

---

### Fixture B — Standard Workflow Entries

**File:** `po/testdata/fixture_b.po`

```po
msgid ""
msgstr ""
"Content-Type: text/plain; charset=UTF-8\n"
"Content-Transfer-Encoding: 8bit\n"
"Language: de\n"
"Plural-Forms: nplurals=2; plural=(n != 1);\n"

#. Button label in the save dialog
#: src/ui/dialogs.c:42
msgid "Save changes"
msgstr "Änderungen speichern"

msgid "Cancel"
msgstr ""

#, fuzzy
#| msgid "Delete item"
msgid "Delete selected items"
msgstr "Element löschen"

#, c-format
msgid "%d item"
msgid_plural "%d items"
msgstr[0] ""
msgstr[1] ""
```

**Expected state after parse:**

| # | msgid | state | notes |
|---|---|---|---|
| 1 | `Save changes` | translated | extracted comment and reference set |
| 2 | `Cancel` | untranslated | |
| 3 | `Delete selected items` | fuzzy | previous msgid = `"Delete item"` |
| 4 | `%d item` | untranslated | plural; flags = `["c-format"]` |

Stats: total=4, translated=1, fuzzy=1, untranslated=2

---

### Fixture C — Full Coverage

**File:** `po/testdata/fixture_c.po`

```po
msgid ""
msgstr ""
"Content-Type: text/plain; charset=UTF-8\n"
"Content-Transfer-Encoding: 8bit\n"
"Language: de-AT\n"
"Plural-Forms: nplurals=2; plural=(n != 1);\n"

# This is a translator comment
#. Extracted from source
#: src/app.c:10 src/app.c:11
#, c-format
msgid "Hello, %s!"
msgstr "Hallo, %s!"

#, fuzzy
#| msgid "Old source"
#| msgid_plural "Old sources"
msgid "New source"
msgid_plural "New sources"
msgstr[0] "Alte Quelle"
msgstr[1] "Alte Quellen"

msgctxt "navigation button"
msgid "Back"
msgstr "Zurück"

msgctxt "navigation button"
msgid "Forward"
msgstr ""

#~ msgid "Removed string"
#~ msgstr "Entfernte Zeichenkette"

msgid "Multi"
"line"
msgstr "Mehr"
"zeilig"
```

**Expected state after parse:**

| # | msgid | msgctxt | state | notes |
|---|---|---|---|---|
| 1 | `Hello, %s!` | — | translated | all comment/flag/reference fields set |
| 2 | `New source` | — | fuzzy | plural; previous msgid and plural set |
| 3 | `Back` | `navigation button` | translated | |
| 4 | `Forward` | `navigation button` | untranslated | |
| 5 | `Multiline` | — | translated | msgid and msgstr joined from continuation lines |

One obsolete node between entry 4 and entry 5, at its original file position.

Language = `"de-AT"`, language name = `"German (Austria)"`

---

### Fixture D — Windows Line Endings

**File:** `po/testdata/fixture_d.po`

Identical content to Fixture A with `\r\n` line endings throughout.

**Expected:** Parses identically to Fixture A. Writer output uses `\n` only.

---

### Fixture E — Missing Plural-Forms, CLDR Fallback

**File:** `po/testdata/fixture_e.po`

```po
msgid ""
msgstr ""
"Content-Type: text/plain; charset=UTF-8\n"
"Language: ja\n"

msgid "item"
msgid_plural "items"
msgstr[0] ""
```

**Expected:** One parse warning about missing `Plural-Forms`. Plural count = `1`
(Japanese per CLDR).

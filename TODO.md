# pollo — Implementation TODO

Ordered by dependency. Complete each section before moving to the next.

---

## 1. Project Scaffolding

- [ ] Create `go.mod` with module path `github.com/yourusername/pollo`, Go 1.26
- [ ] Create `main.go` — entry point only; delegates to `cmd.Execute()`
- [ ] Create directory structure: `cmd/`, `po/`, `po/testdata/`
- [ ] Run `go mod tidy` to fetch `github.com/spf13/cobra`

---

## 2. `po` package — Types (`po/types.go`)

- [ ] Define `Entry` struct with all fields: translator comment, extracted comment, references, flags, previous msgid, previous msgid_plural, msgctxt, msgid, msgid_plural, msgstr, msgstr[n] slice
- [ ] Define `ObsoleteNode` struct (raw lines)
- [ ] Define `Node` interface or tagged union (Entry | ObsoleteNode)
- [ ] Define `File` struct: ordered node list, header entry pointer, nplurals int, language tag string, language display name string
- [ ] Implement `Entry.Key() string` — `msgctxt + U+0004 + msgid` if context present, else `msgid`
- [ ] Implement `Entry.State() string` — returns `"translated"`, `"fuzzy"`, or `"untranslated"`

---

## 3. `po` package — CLDR Plural Rules & Language Names (`po/plural.go`)

- [ ] Embed CLDR nplurals lookup table (all languages from Appendix B, full CLDR coverage)
- [ ] Embed language display name table (all languages listed in Appendix B)
- [ ] Implement `LookupNplurals(tag string) (int, bool)` — case-insensitive, full tag then base language fallback
- [ ] Implement `LookupLanguageName(tag string) string` — full tag first, base language fallback, raw tag if unknown

---

## 4. `po` package — Parser (`po/parser.go`)

- [ ] Implement `Parse(r io.Reader) (*File, []string, error)` returning file, warnings slice, error
- [ ] Strip `\r` from `\r\n` line endings
- [ ] Parse all line types: `msgid`, `msgstr`, `msgstr[n]`, `msgid_plural`, `msgctxt`, `# `, `#.`, `#:`, `#,`, `#|`, `#~`, continuation `"..."` lines
- [ ] Join multi-line quoted strings (continuation lines concatenated after unescaping)
- [ ] Unescape PO sequences: `\n`, `\t`, `\r`, `\\`, `\"`, `\a`, `\b`, `\f`, `\v`; unknown `\X` passes through with warning
- [ ] Parse `#| msgid` and `#| msgid_plural` with continuation line support
- [ ] Identify and store header entry (first non-blank, non-comment entry with empty msgid)
- [ ] Reject (fatal) if `Content-Type` charset is present and not UTF-8 (case-insensitive)
- [ ] Contiguous `#~` lines → single ObsoleteNode at original position
- [ ] Validate `msgstr[n]` indices are contiguous starting from 0; fatal error if gap
- [ ] Plural count resolution (in order): parse `Plural-Forms` header → CLDR fallback (with warning) → fatal if plural entries and unresolvable → default 1 if no plural entries
- [ ] Resolve language display name from tag
- [ ] Detect duplicate keys: retain both entries in order, emit warning
- [ ] Accept empty input as zero entries, no error
- [ ] Accept absent charset header (treat as UTF-8)
- [ ] Handle no header entry gracefully (language and plural-forms unresolvable from header)

---

## 5. `po` package — Line Wrapping (`po/wrap.go`)

- [ ] Implement `WrapValue(keyword, value string, col int) string` — wraps to `col` runes (0 = disabled)
- [ ] Short string fitting on one line → single-line output
- [ ] String exactly at column limit → single-line
- [ ] String with `\n` escapes → always multi-line (keyword `""` on first line)
- [ ] String over limit → multi-line with `keyword ""` first line, then continuation `"..."` lines
- [ ] Split on `\n` sequences first (keep `\n` at end of each segment)
- [ ] Further split long segments at last space before limit
- [ ] No space boundary in segment → emit oversized line without splitting
- [ ] Empty value → `keyword ""\n`
- [ ] Count runes, not bytes (CJK correct)

---

## 6. `po` package — Writer (`po/writer.go`)

- [ ] Implement `Write(w io.Writer, f *File) error`
- [ ] Write header entry first, no preceding blank line
- [ ] Write each subsequent active entry preceded by one blank line
- [ ] Write ObsoleteNodes verbatim at their original position in node list
- [ ] Field order per entry: translator comment, extracted comment, references, flags, `#| msgid`, `#| msgid_plural`, msgctxt, msgid, msgid_plural, msgstr / msgstr[n]
- [ ] Omit empty/nil fields entirely (no empty `#,` line if flags empty)
- [ ] Apply string escaping: `\` → `\\`, `"` → `\"`, newline → `\n`, tab → `\t`, CR → `\r`
- [ ] Apply line wrapping (76-column, rune count) via `wrap.go`
- [ ] File ends with exactly one trailing newline
- [ ] Implement atomic write helper: write to temp file in same dir, `os.Rename` over target; clean up temp on failure

---

## 7. `po` package — Placeholder Validation (`po/validate.go`)

- [ ] Implement `ValidatePlaceholders(entry *Entry, translations []string) (warning string)`
- [ ] Only validate when entry flags contain a known format marker
- [ ] c-format / objc-format / javascript-format / php-format: full POSIX printf regex (`%[flags][width][.prec][len]specifier`, positional `%1$s`)
- [ ] python-format: printf-style + named groups `%(name)s`
- [ ] python-brace-format: `{0}`, `{name}`, `{value:.2f}`
- [ ] sh-format: `$VAR`, `${VAR}`
- [ ] Exclude `%%` from both sides
- [ ] Multiset equality comparison (order-independent, count-sensitive)
- [ ] Warn format: `placeholder mismatch: source has [%d, %s] but translation has [%s]` (sorted)
- [ ] Unknown format flag → no validation, no warning

---

## 8. `po` package — Skip File (`po/skipfile.go`)

- [ ] Implement `ReadSkipFile(path string) (map[string]struct{}, error)` — missing file → empty map, no error
- [ ] Parse lines: blank and `#`-prefixed → skip; key-only or `key\treason` format
- [ ] Unescape key: `\n` → newline, `\r` → CR, `\\` → backslash; U+0004 preserved as-is
- [ ] Ignore keys not found in PO file (handled by caller)

---

## 9. `cmd` package — Root & JSON helpers (`cmd/root.go`)

- [ ] Set up cobra root command with `Execute()` function
- [ ] Implement JSON output helper: use `encoding/json` encoder with `SetEscapeHTML(false)`, always write to stdout, always newline-terminated
- [ ] Implement error output helper: `{"ok": false, "error": "..."}` to stdout, exit 1
- [ ] Implement file-or-inline flag pair helper (e.g. `--id` / `--id-file` mutual exclusion + UTF-8 validation)

---

## 10. `cmd` package — `stats` subcommand (`cmd/stats.go`)

- [ ] Register `pollo stats <file>` command
- [ ] Open and parse the PO file
- [ ] Compute total, translated, fuzzy, untranslated, remaining counts (exclude header and obsolete)
- [ ] Output JSON: `file`, `language`, `language_name`, `total`, `translated`, `fuzzy`, `untranslated`, `remaining`, `parse_warnings`
- [ ] Exit 0 on success, 1 on fatal parse error

---

## 11. `cmd` package — `get` subcommand (`cmd/get.go`)

- [ ] Register `pollo get <file>` command with all flags: `--id`, `--id-file`, `--context`, `--context-file`, `--skip-ids-file`, `--order`, `--include-translated`
- [ ] Enforce `--id`/`--id-file` mutual exclusion with `--order`, `--skip-ids-file`, `--include-translated`
- [ ] Fetch-by-ID path: look up entry by key; exit 2 if not found
- [ ] Iteration path: apply order (fuzzy-first default, untranslated-first), apply skip file, return first qualifying entry
- [ ] `--include-translated`: all non-header, non-obsolete entries in file order, still honouring skip
- [ ] When no qualifying entry: return done sentinel `{"done": true, ...}`
- [ ] Build full response JSON for non-done: all fields always present (null for absent optionals); `flags` never contains `"fuzzy"`; `current_msgstr` vs `current_msgstr_plural` mutually null; `plural_count` null for singular
- [ ] Compute `remaining`, `remaining_fuzzy`, `remaining_untranslated` (skip file respected in counts)
- [ ] Exit 0 (success + done), 1 (fatal error), 2 (entry not found with `--id`)

---

## 12. `cmd` package — `set` subcommand (`cmd/set.go`)

- [ ] Register `pollo set <file>` command with all flags: `--id`, `--id-file`, `--context`, `--context-file`, `--translation`, `--translation-file`, `--translations`, `--translations-file`
- [ ] Enforce `--id`/`--id-file` exactly one required
- [ ] Enforce exactly one of the four translation flags
- [ ] Validate UTF-8 on all string inputs
- [ ] Hard error: entry not found → `{"ok": false, "error": ...}`
- [ ] Hard error: singular entry + plural translation flag
- [ ] Hard error: plural entry + singular translation flag
- [ ] Hard error: plural array length ≠ nplurals
- [ ] Apply translation to entry in memory: set msgstr/msgstr[n], remove `fuzzy` from flags, clear `#|` lines
- [ ] Run placeholder validation; if warning, include `"warning"` field in response
- [ ] Atomic write to file
- [ ] Compute remaining counts from post-write in-memory state
- [ ] Output JSON: `ok`, `file`, `msgid`, `remaining`, `remaining_fuzzy`, `remaining_untranslated`, `parse_warnings`; include `warning` only if raised
- [ ] Exit 0 on success (with or without warning), 1 on any error

---

## 13. Test Fixtures (`po/testdata/`)

- [ ] Write `fixture_a.po` — minimal, header only (from Appendix A)
- [ ] Write `fixture_b.po` — standard workflow: 1 translated, 1 untranslated, 1 fuzzy, 1 plural untranslated (from Appendix A)
- [ ] Write `fixture_c.po` — full coverage: all comment types, fuzzy plural, msgctxt, obsolete block, multi-line (from Appendix A)
- [ ] Write `fixture_d.po` — identical to fixture_a content but with `\r\n` line endings
- [ ] Write `fixture_e.po` — missing Plural-Forms, Japanese, CLDR fallback (from Appendix A)

---

## 14. Tests (`po/po_test.go`)

All tests in package `po_test`, testing exported API only.

### Parser tests
- [ ] Minimal file (fixture_a) — header present, zero entries, language=de, nplurals=2
- [ ] Single fully-translated entry
- [ ] Untranslated entry (empty msgstr)
- [ ] Plural entry with all forms filled
- [ ] Fuzzy entry with `#| msgid`
- [ ] Fuzzy plural entry with `#| msgid_plural`
- [ ] Entry with `msgctxt`
- [ ] Multi-line msgid and msgstr
- [ ] Multi-line `#| msgid` continuation
- [ ] `#~` blocks interleaved between active entries
- [ ] `#~` block at end of file
- [ ] Translator comment (`# `)
- [ ] Extracted comment (`#.`)
- [ ] Source references (`#:`)
- [ ] Multiple flags on one `#,` line
- [ ] Duplicate key → two entries retained + warning emitted
- [ ] Non-UTF-8 charset header → fatal error
- [ ] Absent charset header → accepted
- [ ] Empty input → zero entries, no error
- [ ] Windows `\r\n` line endings → parses identically to `\n` (fixture_d)
- [ ] No header entry
- [ ] Non-contiguous `msgstr[n]` indices → fatal error
- [ ] Unrecognised escape sequence → pass-through + warning

### Entry state tests
- [ ] Singular translated
- [ ] Singular untranslated (empty msgstr)
- [ ] Singular untranslated (whitespace-only msgstr)
- [ ] Fuzzy with non-empty msgstr → state is "fuzzy", not "translated"
- [ ] Plural all forms filled → translated
- [ ] Plural one form empty → untranslated

### Writer tests
- [ ] Parse then write minimal file; output is valid PO
- [ ] Parse fixture_c, write, re-parse; all fields round-trip correctly
- [ ] After `set` equivalent: fuzzy flag removed, `#|` lines absent, `#,` omitted if no remaining flags
- [ ] Obsolete blocks appear at their original position (not moved to end)

### Line wrapping tests
- [ ] Short string → single-line output
- [ ] String exactly at column limit → single-line
- [ ] String one rune over limit → multi-line
- [ ] String containing `\n` sequences → multi-line always
- [ ] Long token with no space boundary → oversized line emitted without splitting
- [ ] CJK characters counted as one rune each (not bytes)
- [ ] Empty value → `keyword ""\n`
- [ ] `col = 0` → wrapping disabled

### Placeholder validation tests
- [ ] Matching `c-format` placeholders → no warning
- [ ] Mismatched `c-format` → warning with correct message
- [ ] Positional args `%1$s %2$d` handled correctly
- [ ] `%%` excluded from both sides
- [ ] Python brace-format match → no warning
- [ ] Python brace-format mismatch → warning
- [ ] Unrecognised format flag → no warning
- [ ] No format flag → no warning

### Plural count resolution tests
- [ ] Resolved from `Plural-Forms` header
- [ ] CLDR fallback for German (`de` → 2) with warning
- [ ] CLDR fallback for Arabic (`ar` → 6) with warning
- [ ] CLDR fallback for Japanese (`ja` → 1) with warning (fixture_e)
- [ ] Language absent + plural entries → fatal error
- [ ] Language absent + no plural entries → nplurals=1, no error

### Skip file tests
- [ ] Empty file → empty map
- [ ] Simple msgid entries parsed correctly
- [ ] Key with `msgctxt` U+0004 separator preserved
- [ ] Key containing escaped newline (`\n` two-char → actual newline)
- [ ] Entry with tab-separated reason → key extracted without reason
- [ ] Missing file → empty map, no error

### Integration test (fixture_b)
- [ ] Parse fixture_b; assert stats: total=4, translated=1, fuzzy=1, untranslated=2
- [ ] Get first entry (fuzzy-first order) → assert it is the fuzzy entry ("Delete selected items")
- [ ] Apply set with translation → assert fuzzy flag removed, `#|` lines cleared
- [ ] Loop get/set until done; write to buffer, re-parse, assert all 4 entries translated

---

## 15. Final Validation

- [ ] `go vet ./...` — no warnings
- [ ] `go test ./...` — all tests pass
- [ ] `go build -o pollo .` — binary builds successfully
- [ ] Manual smoke test: `pollo stats`, `pollo get`, `pollo set` against a real PO file
- [ ] Verify JSON output never HTML-escapes `<`, `>`, `&` in string values
- [ ] Verify atomic write: temp file cleaned up on error, original untouched

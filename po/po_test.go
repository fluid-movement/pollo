package po_test

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/fluid-movement/pollo/po"
)

// helper to open a fixture file
func openFixture(t *testing.T, name string) *os.File {
	t.Helper()
	f, err := os.Open("testdata/" + name)
	if err != nil {
		t.Fatalf("opening fixture %s: %v", name, err)
	}
	t.Cleanup(func() { f.Close() })
	return f
}

func mustParse(t *testing.T, r interface{ Read([]byte) (int, error) }) (*po.File, []string) {
	t.Helper()
	file, warns, err := po.Parse(r)
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}
	return file, warns
}

func activeEntries(f *po.File) []*po.Entry {
	var entries []*po.Entry
	for _, node := range f.Nodes {
		if e, ok := node.(*po.Entry); ok {
			entries = append(entries, e)
		}
	}
	return entries
}

// ---- Parser tests ----

func TestParserMinimalFile(t *testing.T) {
	file, _ := mustParse(t, openFixture(t, "fixture_a.po"))
	if file.Header == nil {
		t.Fatal("expected header entry")
	}
	if file.Language != "de" {
		t.Errorf("language = %q, want %q", file.Language, "de")
	}
	if file.Nplurals != 2 {
		t.Errorf("nplurals = %d, want 2", file.Nplurals)
	}
	entries := activeEntries(file)
	if len(entries) != 0 {
		t.Errorf("expected 0 entries, got %d", len(entries))
	}
}

func TestParserSingleTranslated(t *testing.T) {
	src := `msgid ""
msgstr ""
"Content-Type: text/plain; charset=UTF-8\n"
"Language: de\n"
"Plural-Forms: nplurals=2; plural=(n != 1);\n"

msgid "Hello"
msgstr "Hallo"
`
	file, _ := mustParse(t, strings.NewReader(src))
	entries := activeEntries(file)
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	e := entries[0]
	if e.Msgid != "Hello" {
		t.Errorf("msgid = %q", e.Msgid)
	}
	if e.Msgstr != "Hallo" {
		t.Errorf("msgstr = %q", e.Msgstr)
	}
	if e.State() != "translated" {
		t.Errorf("state = %q", e.State())
	}
}

func TestParserUntranslated(t *testing.T) {
	src := `msgid ""
msgstr ""
"Content-Type: text/plain; charset=UTF-8\n"
"Language: de\n"
"Plural-Forms: nplurals=2; plural=(n != 1);\n"

msgid "Hello"
msgstr ""
`
	file, _ := mustParse(t, strings.NewReader(src))
	entries := activeEntries(file)
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].State() != "untranslated" {
		t.Errorf("state = %q, want untranslated", entries[0].State())
	}
}

func TestParserPluralAllForms(t *testing.T) {
	src := `msgid ""
msgstr ""
"Content-Type: text/plain; charset=UTF-8\n"
"Language: de\n"
"Plural-Forms: nplurals=2; plural=(n != 1);\n"

msgid "%d item"
msgid_plural "%d items"
msgstr[0] "%d Element"
msgstr[1] "%d Elemente"
`
	file, _ := mustParse(t, strings.NewReader(src))
	entries := activeEntries(file)
	if len(entries) != 1 {
		t.Fatalf("expected 1, got %d", len(entries))
	}
	e := entries[0]
	if e.MsgidPlural != "%d items" {
		t.Errorf("msgid_plural = %q", e.MsgidPlural)
	}
	if len(e.MsgstrPlural) != 2 {
		t.Fatalf("expected 2 plural forms, got %d", len(e.MsgstrPlural))
	}
	if e.MsgstrPlural[0] != "%d Element" {
		t.Errorf("msgstr[0] = %q", e.MsgstrPlural[0])
	}
	if e.State() != "translated" {
		t.Errorf("state = %q", e.State())
	}
}

func TestParserFuzzyWithPrevMsgid(t *testing.T) {
	src := `msgid ""
msgstr ""
"Content-Type: text/plain; charset=UTF-8\n"
"Language: de\n"
"Plural-Forms: nplurals=2; plural=(n != 1);\n"

#, fuzzy
#| msgid "Delete item"
msgid "Delete selected items"
msgstr "Element löschen"
`
	file, _ := mustParse(t, strings.NewReader(src))
	entries := activeEntries(file)
	if len(entries) != 1 {
		t.Fatalf("expected 1, got %d", len(entries))
	}
	e := entries[0]
	if e.PrevMsgid != "Delete item" {
		t.Errorf("PrevMsgid = %q", e.PrevMsgid)
	}
	if e.State() != "fuzzy" {
		t.Errorf("state = %q", e.State())
	}
}

func TestParserFuzzyPluralWithPrevMsgidPlural(t *testing.T) {
	src := `msgid ""
msgstr ""
"Content-Type: text/plain; charset=UTF-8\n"
"Language: de\n"
"Plural-Forms: nplurals=2; plural=(n != 1);\n"

#, fuzzy
#| msgid "Old source"
#| msgid_plural "Old sources"
msgid "New source"
msgid_plural "New sources"
msgstr[0] "Alte Quelle"
msgstr[1] "Alte Quellen"
`
	file, _ := mustParse(t, strings.NewReader(src))
	entries := activeEntries(file)
	if len(entries) != 1 {
		t.Fatalf("expected 1, got %d", len(entries))
	}
	e := entries[0]
	if e.PrevMsgid != "Old source" {
		t.Errorf("PrevMsgid = %q", e.PrevMsgid)
	}
	if e.PrevMsgidPlural != "Old sources" {
		t.Errorf("PrevMsgidPlural = %q", e.PrevMsgidPlural)
	}
}

func TestParserMsgctxt(t *testing.T) {
	src := `msgid ""
msgstr ""
"Content-Type: text/plain; charset=UTF-8\n"
"Language: de\n"
"Plural-Forms: nplurals=2; plural=(n != 1);\n"

msgctxt "navigation button"
msgid "Back"
msgstr "Zurück"
`
	file, _ := mustParse(t, strings.NewReader(src))
	entries := activeEntries(file)
	if len(entries) != 1 {
		t.Fatalf("expected 1, got %d", len(entries))
	}
	e := entries[0]
	if e.Msgctxt != "navigation button" {
		t.Errorf("Msgctxt = %q", e.Msgctxt)
	}
	if e.Key() != "navigation button\x04Back" {
		t.Errorf("Key = %q", e.Key())
	}
}

func TestParserMultilineMsg(t *testing.T) {
	src := `msgid ""
msgstr ""
"Content-Type: text/plain; charset=UTF-8\n"
"Language: de\n"
"Plural-Forms: nplurals=2; plural=(n != 1);\n"

msgid "Multi"
"line"
msgstr "Mehr"
"zeilig"
`
	file, _ := mustParse(t, strings.NewReader(src))
	entries := activeEntries(file)
	if len(entries) != 1 {
		t.Fatalf("expected 1, got %d", len(entries))
	}
	e := entries[0]
	if e.Msgid != "Multiline" {
		t.Errorf("Msgid = %q, want %q", e.Msgid, "Multiline")
	}
	if e.Msgstr != "Mehrzeilig" {
		t.Errorf("Msgstr = %q, want %q", e.Msgstr, "Mehrzeilig")
	}
}

func TestParserMultilinePrevMsgidContinuation(t *testing.T) {
	src := `msgid ""
msgstr ""
"Content-Type: text/plain; charset=UTF-8\n"
"Language: de\n"
"Plural-Forms: nplurals=2; plural=(n != 1);\n"

#, fuzzy
#| msgid "Old"
" continued"
msgid "New"
msgstr "Neu"
`
	file, _ := mustParse(t, strings.NewReader(src))
	entries := activeEntries(file)
	if len(entries) != 1 {
		t.Fatalf("expected 1, got %d", len(entries))
	}
	e := entries[0]
	if e.PrevMsgid != "Old continued" {
		t.Errorf("PrevMsgid = %q, want %q", e.PrevMsgid, "Old continued")
	}
}

func TestParserObsoleteInterleaved(t *testing.T) {
	file, _ := mustParse(t, openFixture(t, "fixture_c.po"))
	entries := activeEntries(file)
	if len(entries) != 5 {
		t.Fatalf("expected 5 active entries, got %d", len(entries))
	}
	// Check obsolete node position
	obsCount := 0
	obsPos := -1
	for i, node := range file.Nodes {
		if _, ok := node.(*po.ObsoleteNode); ok {
			obsCount++
			obsPos = i
		}
	}
	if obsCount != 1 {
		t.Errorf("expected 1 obsolete node, got %d", obsCount)
	}
	// Obsolete should be between entry 4 (Forward) and entry 5 (Multi)
	if obsPos < 0 {
		t.Error("obsolete node not found")
	}
}

func TestParserObsoleteAtEnd(t *testing.T) {
	src := `msgid ""
msgstr ""
"Content-Type: text/plain; charset=UTF-8\n"
"Language: de\n"
"Plural-Forms: nplurals=2; plural=(n != 1);\n"

msgid "Hello"
msgstr "Hallo"

#~ msgid "Old"
#~ msgstr "Alt"
`
	file, _ := mustParse(t, strings.NewReader(src))
	entries := activeEntries(file)
	if len(entries) != 1 {
		t.Fatalf("expected 1, got %d", len(entries))
	}
	// Last node should be obsolete
	last := file.Nodes[len(file.Nodes)-1]
	if _, ok := last.(*po.ObsoleteNode); !ok {
		t.Error("last node should be obsolete")
	}
}

func TestParserTranslatorComment(t *testing.T) {
	src := `msgid ""
msgstr ""
"Content-Type: text/plain; charset=UTF-8\n"
"Language: de\n"
"Plural-Forms: nplurals=2; plural=(n != 1);\n"

# This is a translator comment
msgid "Hello"
msgstr "Hallo"
`
	file, _ := mustParse(t, strings.NewReader(src))
	entries := activeEntries(file)
	if entries[0].TranslatorComment != "This is a translator comment" {
		t.Errorf("TranslatorComment = %q", entries[0].TranslatorComment)
	}
}

func TestParserExtractedComment(t *testing.T) {
	src := `msgid ""
msgstr ""
"Content-Type: text/plain; charset=UTF-8\n"
"Language: de\n"
"Plural-Forms: nplurals=2; plural=(n != 1);\n"

#. Extracted comment
msgid "Hello"
msgstr "Hallo"
`
	file, _ := mustParse(t, strings.NewReader(src))
	entries := activeEntries(file)
	if entries[0].ExtractedComment != "Extracted comment" {
		t.Errorf("ExtractedComment = %q", entries[0].ExtractedComment)
	}
}

func TestParserSourceReferences(t *testing.T) {
	src := `msgid ""
msgstr ""
"Content-Type: text/plain; charset=UTF-8\n"
"Language: de\n"
"Plural-Forms: nplurals=2; plural=(n != 1);\n"

#: src/app.c:10 src/app.c:11
msgid "Hello"
msgstr "Hallo"
`
	file, _ := mustParse(t, strings.NewReader(src))
	entries := activeEntries(file)
	refs := entries[0].References
	if len(refs) != 2 {
		t.Fatalf("expected 2 refs, got %d", len(refs))
	}
	if refs[0] != "src/app.c:10" || refs[1] != "src/app.c:11" {
		t.Errorf("refs = %v", refs)
	}
}

func TestParserMultipleFlags(t *testing.T) {
	src := `msgid ""
msgstr ""
"Content-Type: text/plain; charset=UTF-8\n"
"Language: de\n"
"Plural-Forms: nplurals=2; plural=(n != 1);\n"

#, fuzzy, c-format
msgid "Hello %s"
msgstr "Hallo %s"
`
	file, _ := mustParse(t, strings.NewReader(src))
	entries := activeEntries(file)
	flags := entries[0].Flags
	if len(flags) != 2 {
		t.Fatalf("expected 2 flags, got %d: %v", len(flags), flags)
	}
}

func TestParserDuplicateKey(t *testing.T) {
	src := `msgid ""
msgstr ""
"Content-Type: text/plain; charset=UTF-8\n"
"Language: de\n"
"Plural-Forms: nplurals=2; plural=(n != 1);\n"

msgid "Hello"
msgstr "Hallo"

msgid "Hello"
msgstr "Hallo2"
`
	file, warns, err := po.Parse(strings.NewReader(src))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	entries := activeEntries(file)
	if len(entries) != 2 {
		t.Errorf("expected 2 entries (both retained), got %d", len(entries))
	}
	if len(warns) == 0 {
		t.Error("expected duplicate key warning")
	}
}

func TestParserNonUTF8Charset(t *testing.T) {
	src := `msgid ""
msgstr ""
"Content-Type: text/plain; charset=ISO-8859-1\n"
"Language: de\n"
`
	_, _, err := po.Parse(strings.NewReader(src))
	if err == nil {
		t.Error("expected fatal error for non-UTF-8 charset")
	}
}

func TestParserAbsentCharset(t *testing.T) {
	src := `msgid ""
msgstr ""
"Language: de\n"
"Plural-Forms: nplurals=2; plural=(n != 1);\n"

msgid "Hello"
msgstr "Hallo"
`
	_, _, err := po.Parse(strings.NewReader(src))
	if err != nil {
		t.Errorf("unexpected error for absent charset: %v", err)
	}
}

func TestParserEmptyInput(t *testing.T) {
	file, warns, err := po.Parse(strings.NewReader(""))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(warns) != 0 {
		t.Errorf("unexpected warnings: %v", warns)
	}
	if file.Header != nil {
		t.Error("expected no header")
	}
	entries := activeEntries(file)
	if len(entries) != 0 {
		t.Errorf("expected 0 entries, got %d", len(entries))
	}
}

func TestParserWindowsLineEndings(t *testing.T) {
	fileA, _ := mustParse(t, openFixture(t, "fixture_a.po"))
	fileD, _ := mustParse(t, openFixture(t, "fixture_d.po"))

	if fileA.Language != fileD.Language {
		t.Errorf("language differs: %q vs %q", fileA.Language, fileD.Language)
	}
	if fileA.Nplurals != fileD.Nplurals {
		t.Errorf("nplurals differs: %d vs %d", fileA.Nplurals, fileD.Nplurals)
	}
}

func TestParserNoHeaderEntry(t *testing.T) {
	src := `msgid "Hello"
msgstr "Hallo"
`
	file, _, err := po.Parse(strings.NewReader(src))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if file.Header != nil {
		t.Error("expected no header")
	}
}

func TestParserNonContiguousMsgstrN(t *testing.T) {
	src := `msgid ""
msgstr ""
"Content-Type: text/plain; charset=UTF-8\n"
"Language: de\n"
"Plural-Forms: nplurals=3; plural=(n != 1);\n"

msgid "%d item"
msgid_plural "%d items"
msgstr[0] "form0"
msgstr[2] "form2"
`
	_, _, err := po.Parse(strings.NewReader(src))
	if err == nil {
		t.Error("expected fatal error for non-contiguous msgstr[n]")
	}
}

func TestParserUnrecognisedEscape(t *testing.T) {
	src := `msgid ""
msgstr ""
"Content-Type: text/plain; charset=UTF-8\n"
"Language: de\n"
"Plural-Forms: nplurals=2; plural=(n != 1);\n"

msgid "Hello\q world"
msgstr "Hallo\q welt"
`
	file, warns, err := po.Parse(strings.NewReader(src))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(warns) == 0 {
		t.Error("expected warning for unrecognised escape")
	}
	entries := activeEntries(file)
	// \q should pass through as-is
	if !strings.Contains(entries[0].Msgid, `\q`) {
		t.Errorf("expected \\q in msgid, got %q", entries[0].Msgid)
	}
}

// ---- Entry state tests ----

func TestEntryStateSingularTranslated(t *testing.T) {
	e := &po.Entry{Msgid: "Hello", Msgstr: "Hallo"}
	if e.State() != "translated" {
		t.Errorf("state = %q", e.State())
	}
}

func TestEntryStateSingularUntranslatedEmpty(t *testing.T) {
	e := &po.Entry{Msgid: "Hello", Msgstr: ""}
	if e.State() != "untranslated" {
		t.Errorf("state = %q", e.State())
	}
}

func TestEntryStateSingularUntranslatedWhitespace(t *testing.T) {
	e := &po.Entry{Msgid: "Hello", Msgstr: "   \t\n"}
	if e.State() != "untranslated" {
		t.Errorf("state = %q", e.State())
	}
}

func TestEntryStateFuzzyWithNonEmptyMsgstr(t *testing.T) {
	e := &po.Entry{Msgid: "Hello", Msgstr: "Hallo", Flags: []string{"fuzzy"}}
	if e.State() != "fuzzy" {
		t.Errorf("state = %q, want fuzzy", e.State())
	}
}

func TestEntryStatePluralAllFilled(t *testing.T) {
	e := &po.Entry{Msgid: "item", MsgidPlural: "items", MsgstrPlural: []string{"Element", "Elemente"}}
	if e.State() != "translated" {
		t.Errorf("state = %q", e.State())
	}
}

func TestEntryStatePluralOneEmpty(t *testing.T) {
	e := &po.Entry{Msgid: "item", MsgidPlural: "items", MsgstrPlural: []string{"Element", ""}}
	if e.State() != "untranslated" {
		t.Errorf("state = %q", e.State())
	}
}

// ---- Writer tests ----

func TestWriterMinimalFile(t *testing.T) {
	file, _ := mustParse(t, openFixture(t, "fixture_a.po"))
	var buf bytes.Buffer
	if err := po.Write(&buf, file); err != nil {
		t.Fatalf("write error: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, `msgid ""`) {
		t.Error("output missing msgid")
	}
	if !strings.Contains(out, `msgstr ""`) {
		t.Error("output missing msgstr")
	}
	// Should be valid PO (re-parseable)
	file2, _, err := po.Parse(strings.NewReader(out))
	if err != nil {
		t.Fatalf("re-parse error: %v", err)
	}
	if file2.Language != "de" {
		t.Errorf("language = %q", file2.Language)
	}
}

func TestWriterRoundTripFixtureC(t *testing.T) {
	orig, _ := mustParse(t, openFixture(t, "fixture_c.po"))
	var buf bytes.Buffer
	if err := po.Write(&buf, orig); err != nil {
		t.Fatalf("write error: %v", err)
	}
	reparsed, _, err := po.Parse(strings.NewReader(buf.String()))
	if err != nil {
		t.Fatalf("re-parse error: %v", err)
	}

	origEntries := activeEntries(orig)
	reparsedEntries := activeEntries(reparsed)
	if len(origEntries) != len(reparsedEntries) {
		t.Fatalf("entry count: orig=%d, reparsed=%d", len(origEntries), len(reparsedEntries))
	}
	for i := range origEntries {
		o := origEntries[i]
		r := reparsedEntries[i]
		if o.Msgid != r.Msgid {
			t.Errorf("entry %d: msgid %q != %q", i, o.Msgid, r.Msgid)
		}
		if o.Msgstr != r.Msgstr {
			t.Errorf("entry %d: msgstr %q != %q", i, o.Msgstr, r.Msgstr)
		}
		if o.Msgctxt != r.Msgctxt {
			t.Errorf("entry %d: msgctxt %q != %q", i, o.Msgctxt, r.Msgctxt)
		}
		if o.MsgidPlural != r.MsgidPlural {
			t.Errorf("entry %d: msgid_plural %q != %q", i, o.MsgidPlural, r.MsgidPlural)
		}
		if o.PrevMsgid != r.PrevMsgid {
			t.Errorf("entry %d: prev_msgid %q != %q", i, o.PrevMsgid, r.PrevMsgid)
		}
		if o.TranslatorComment != r.TranslatorComment {
			t.Errorf("entry %d: translator_comment %q != %q", i, o.TranslatorComment, r.TranslatorComment)
		}
	}
}

func TestWriterAfterSetFuzzyRemovedAndPrevCleared(t *testing.T) {
	src := `msgid ""
msgstr ""
"Content-Type: text/plain; charset=UTF-8\n"
"Language: de\n"
"Plural-Forms: nplurals=2; plural=(n != 1);\n"

#, fuzzy, c-format
#| msgid "Old"
msgid "New"
msgstr "Alt"
`
	file, _, err := po.Parse(strings.NewReader(src))
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	entries := activeEntries(file)
	entries[0].Msgstr = "Neu"
	entries[0].Flags = removeFlag(entries[0].Flags, "fuzzy")
	entries[0].PrevMsgid = ""

	var buf bytes.Buffer
	if err := po.Write(&buf, file); err != nil {
		t.Fatalf("write error: %v", err)
	}
	out := buf.String()
	if strings.Contains(out, "fuzzy") {
		t.Error("output should not contain fuzzy flag")
	}
	if strings.Contains(out, "#| msgid") {
		t.Error("output should not contain previous msgid")
	}
	// c-format flag should still be present
	if !strings.Contains(out, "c-format") {
		t.Error("c-format flag should be preserved")
	}
	// #, line should still be present since c-format remains
	if !strings.Contains(out, "#,") {
		t.Error("#, line should still be present for c-format")
	}
}

func TestWriterObsoleteAtOriginalPosition(t *testing.T) {
	file, _ := mustParse(t, openFixture(t, "fixture_c.po"))
	var buf bytes.Buffer
	if err := po.Write(&buf, file); err != nil {
		t.Fatalf("write error: %v", err)
	}
	out := buf.String()
	// "Removed string" should appear before "Multi"
	obsIdx := strings.Index(out, "Removed string")
	multiIdx := strings.Index(out, `msgid "Multi`)
	if obsIdx < 0 {
		t.Fatal("obsolete entry not found in output")
	}
	if multiIdx < 0 {
		t.Fatal("Multi entry not found in output")
	}
	if obsIdx > multiIdx {
		t.Error("obsolete node should appear before Multi entry")
	}
}

func removeFlag(flags []string, flag string) []string {
	var result []string
	for _, f := range flags {
		if f != flag {
			result = append(result, f)
		}
	}
	return result
}

// ---- Line wrapping tests ----

func TestWrapShortString(t *testing.T) {
	out := po.WrapValue("msgstr", "Hello", 76)
	want := `msgstr "Hello"` + "\n"
	if out != want {
		t.Errorf("wrap = %q, want %q", out, want)
	}
}

func TestWrapExactlyAtLimit(t *testing.T) {
	// keyword "value" = 76 chars: "msgstr " (7) + '"' (1) + value + '"' (1) = 76
	// value length = 76 - 9 = 67
	value := strings.Repeat("x", 67)
	out := po.WrapValue("msgstr", value, 76)
	// Should be single line
	lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
	if len(lines) != 1 {
		t.Errorf("expected 1 line, got %d", len(lines))
	}
}

func TestWrapOneRuneOverLimit(t *testing.T) {
	// 68 x's → over limit of 67 → multi-line
	value := strings.Repeat("x", 68)
	out := po.WrapValue("msgstr", value, 76)
	if !strings.HasPrefix(out, `msgstr ""`) {
		t.Errorf("expected multi-line, got: %q", out)
	}
}

func TestWrapWithNewlineEscapeAlwaysMultiLine(t *testing.T) {
	out := po.WrapValue("msgstr", `Hello\nWorld`, 76)
	if !strings.HasPrefix(out, `msgstr ""`) {
		t.Errorf("expected multi-line for \\n, got: %q", out)
	}
}

func TestWrapLongTokenNoSpace(t *testing.T) {
	// A URL with no spaces - should emit oversized line
	value := strings.Repeat("a", 200)
	out := po.WrapValue("msgstr", value, 76)
	// Should not panic; should have content
	if !strings.Contains(out, value[:10]) {
		t.Error("expected content in output")
	}
}

func TestWrapCJKRunes(t *testing.T) {
	// Each CJK char is 3 bytes but 1 rune
	// 67 CJK chars should fit in 76 col limit (9 for keyword + quotes)
	value := strings.Repeat("中", 67)
	out := po.WrapValue("msgstr", value, 76)
	lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
	if len(lines) != 1 {
		t.Errorf("expected single line for 67 CJK chars, got %d lines", len(lines))
	}
}

func TestWrapEmpty(t *testing.T) {
	out := po.WrapValue("msgstr", "", 76)
	want := `msgstr ""` + "\n"
	if out != want {
		t.Errorf("wrap empty = %q, want %q", out, want)
	}
}

func TestWrapColZeroDisabled(t *testing.T) {
	value := strings.Repeat("x", 200)
	out := po.WrapValue("msgstr", value, 0)
	lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
	if len(lines) != 1 {
		t.Errorf("col=0 should disable wrapping, got %d lines", len(lines))
	}
}

// ---- Placeholder validation tests ----

func TestValidateCFormatMatch(t *testing.T) {
	e := &po.Entry{Msgid: "Hello, %s!", Flags: []string{"c-format"}}
	w := po.ValidatePlaceholders(e, []string{"Hallo, %s!"})
	if w != "" {
		t.Errorf("expected no warning, got: %q", w)
	}
}

func TestValidateCFormatMismatch(t *testing.T) {
	e := &po.Entry{Msgid: "Hello, %s! You have %d messages.", Flags: []string{"c-format"}}
	w := po.ValidatePlaceholders(e, []string{"Hallo, %s!"})
	if w == "" {
		t.Error("expected mismatch warning")
	}
	if !strings.Contains(w, "placeholder mismatch") {
		t.Errorf("warning = %q", w)
	}
}

func TestValidatePositionalArgs(t *testing.T) {
	e := &po.Entry{Msgid: "%1$s has %2$d items", Flags: []string{"c-format"}}
	w := po.ValidatePlaceholders(e, []string{"%2$d items belongs to %1$s"})
	if w != "" {
		t.Errorf("expected no warning for positional args, got: %q", w)
	}
}

func TestValidatePercentPercentExcluded(t *testing.T) {
	e := &po.Entry{Msgid: "100%% done", Flags: []string{"c-format"}}
	w := po.ValidatePlaceholders(e, []string{"100%% fertig"})
	if w != "" {
		t.Errorf("expected no warning, got: %q", w)
	}
}

func TestValidatePythonBraceFormatMatch(t *testing.T) {
	e := &po.Entry{Msgid: "Hello, {name}!", Flags: []string{"python-brace-format"}}
	w := po.ValidatePlaceholders(e, []string{"Hallo, {name}!"})
	if w != "" {
		t.Errorf("expected no warning, got: %q", w)
	}
}

func TestValidatePythonBraceFormatMismatch(t *testing.T) {
	e := &po.Entry{Msgid: "Hello, {name}!", Flags: []string{"python-brace-format"}}
	w := po.ValidatePlaceholders(e, []string{"Hallo!"})
	if w == "" {
		t.Error("expected mismatch warning")
	}
}

func TestValidateUnrecognisedFormat(t *testing.T) {
	e := &po.Entry{Msgid: "Hello %s", Flags: []string{"some-unknown-format"}}
	w := po.ValidatePlaceholders(e, []string{"Hallo"})
	if w != "" {
		t.Errorf("expected no warning for unknown format, got: %q", w)
	}
}

func TestValidateNoFormat(t *testing.T) {
	e := &po.Entry{Msgid: "Hello %s", Flags: []string{}}
	w := po.ValidatePlaceholders(e, []string{"Hallo"})
	if w != "" {
		t.Errorf("expected no warning when no format flag, got: %q", w)
	}
}

// ---- Plural count resolution tests ----

func TestPluralFromHeader(t *testing.T) {
	src := `msgid ""
msgstr ""
"Content-Type: text/plain; charset=UTF-8\n"
"Language: de\n"
"Plural-Forms: nplurals=2; plural=(n != 1);\n"
`
	file, _, err := po.Parse(strings.NewReader(src))
	if err != nil {
		t.Fatal(err)
	}
	if file.Nplurals != 2 {
		t.Errorf("nplurals = %d, want 2", file.Nplurals)
	}
}

func TestPluralCLDRFallbackGerman(t *testing.T) {
	src := `msgid ""
msgstr ""
"Content-Type: text/plain; charset=UTF-8\n"
"Language: de\n"

msgid "item"
msgid_plural "items"
msgstr[0] ""
msgstr[1] ""
`
	file, warns, err := po.Parse(strings.NewReader(src))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if file.Nplurals != 2 {
		t.Errorf("nplurals = %d, want 2", file.Nplurals)
	}
	if len(warns) == 0 {
		t.Error("expected CLDR fallback warning")
	}
}

func TestPluralCLDRFallbackArabic(t *testing.T) {
	src := `msgid ""
msgstr ""
"Content-Type: text/plain; charset=UTF-8\n"
"Language: ar\n"

msgid "item"
msgid_plural "items"
msgstr[0] ""
msgstr[1] ""
`
	file, _, err := po.Parse(strings.NewReader(src))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if file.Nplurals != 6 {
		t.Errorf("nplurals = %d, want 6", file.Nplurals)
	}
}

func TestPluralCLDRFallbackJapanese(t *testing.T) {
	file, warns, err := po.Parse(openFixture(t, "fixture_e.po"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if file.Nplurals != 1 {
		t.Errorf("nplurals = %d, want 1", file.Nplurals)
	}
	if len(warns) == 0 {
		t.Error("expected CLDR fallback warning for fixture_e")
	}
}

func TestPluralLanguageAbsentWithPluralEntries(t *testing.T) {
	src := `msgid ""
msgstr ""
"Content-Type: text/plain; charset=UTF-8\n"

msgid "item"
msgid_plural "items"
msgstr[0] ""
msgstr[1] ""
`
	_, _, err := po.Parse(strings.NewReader(src))
	if err == nil {
		t.Error("expected fatal error: language absent with plural entries")
	}
}

func TestPluralLanguageAbsentNoPluralEntries(t *testing.T) {
	src := `msgid ""
msgstr ""
"Content-Type: text/plain; charset=UTF-8\n"

msgid "Hello"
msgstr ""
`
	file, _, err := po.Parse(strings.NewReader(src))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if file.Nplurals != 1 {
		t.Errorf("nplurals = %d, want 1", file.Nplurals)
	}
}

// ---- Skip file tests ----

func TestSkipFileEmpty(t *testing.T) {
	tmp, err := os.CreateTemp(t.TempDir(), "skip*.txt")
	if err != nil {
		t.Fatal(err)
	}
	tmp.Close()

	m, err := po.ReadSkipFile(tmp.Name())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(m) != 0 {
		t.Errorf("expected empty map, got %v", m)
	}
}

func TestSkipFileSimpleEntries(t *testing.T) {
	content := "Hello\nWorld\n"
	path := writeTempFile(t, content)
	m, err := po.ReadSkipFile(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := m["Hello"]; !ok {
		t.Error("expected Hello in skip map")
	}
	if _, ok := m["World"]; !ok {
		t.Error("expected World in skip map")
	}
}

func TestSkipFileWithMsgctxtSeparator(t *testing.T) {
	key := "navigation button\x04Back"
	content := key + "\n"
	path := writeTempFile(t, content)
	m, err := po.ReadSkipFile(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := m[key]; !ok {
		t.Errorf("expected key with EOT separator, map = %v", m)
	}
}

func TestSkipFileEscapedNewline(t *testing.T) {
	// Key with escaped newline: \n (two chars) → actual newline
	content := `Some string with\nnewline` + "\n"
	path := writeTempFile(t, content)
	m, err := po.ReadSkipFile(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := "Some string with\nnewline"
	if _, ok := m[expected]; !ok {
		t.Errorf("expected key %q in map, got %v", expected, m)
	}
}

func TestSkipFileWithReason(t *testing.T) {
	content := "INTERNAL_KEY\tNot user-visible\n"
	path := writeTempFile(t, content)
	m, err := po.ReadSkipFile(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := m["INTERNAL_KEY"]; !ok {
		t.Error("expected INTERNAL_KEY in skip map")
	}
	if _, ok := m["INTERNAL_KEY\tNot user-visible"]; ok {
		t.Error("should not have full line as key")
	}
}

func TestSkipFileMissing(t *testing.T) {
	m, err := po.ReadSkipFile("/tmp/does-not-exist-pollo-test.txt")
	if err != nil {
		t.Fatalf("unexpected error for missing file: %v", err)
	}
	if len(m) != 0 {
		t.Errorf("expected empty map, got %v", m)
	}
}

func writeTempFile(t *testing.T, content string) string {
	t.Helper()
	tmp, err := os.CreateTemp(t.TempDir(), "skip*.txt")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := tmp.WriteString(content); err != nil {
		t.Fatal(err)
	}
	tmp.Close()
	return tmp.Name()
}

// ---- Integration test (fixture_b) ----

func TestIntegrationFixtureB(t *testing.T) {
	// 1. Parse fixture_b and assert initial stats
	file, _, err := po.Parse(openFixture(t, "fixture_b.po"))
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	entries := activeEntries(file)
	if len(entries) != 4 {
		t.Fatalf("expected 4 entries, got %d", len(entries))
	}

	total, translated, fuzzy, untranslated := 0, 0, 0, 0
	for _, e := range entries {
		switch e.State() {
		case "translated":
			translated++
		case "fuzzy":
			fuzzy++
		case "untranslated":
			untranslated++
		}
		total++
	}
	if total != 4 || translated != 1 || fuzzy != 1 || untranslated != 2 {
		t.Errorf("stats: total=%d translated=%d fuzzy=%d untranslated=%d", total, translated, fuzzy, untranslated)
	}

	// 2. Get first entry with fuzzy-first ordering
	// Fuzzy entries first, then untranslated
	var fuzzyFirst *po.Entry
	for _, e := range entries {
		if e.State() == "fuzzy" {
			fuzzyFirst = e
			break
		}
	}
	if fuzzyFirst == nil {
		t.Fatal("expected fuzzy entry")
	}
	if fuzzyFirst.Msgid != "Delete selected items" {
		t.Errorf("first fuzzy entry = %q, want %q", fuzzyFirst.Msgid, "Delete selected items")
	}

	// 3. Apply set: clear fuzzy flag, set translation, clear #| lines
	fuzzyFirst.Msgstr = "Ausgewählte Elemente löschen"
	fuzzyFirst.Flags = removeFlag(fuzzyFirst.Flags, "fuzzy")
	fuzzyFirst.PrevMsgid = ""

	if fuzzyFirst.State() != "translated" {
		t.Errorf("after set, state = %q, want translated", fuzzyFirst.State())
	}
	if fuzzyFirst.PrevMsgid != "" {
		t.Error("PrevMsgid should be cleared")
	}

	// 4. Loop get/set until done (simulate full workflow)
	// Set all remaining untranslated
	for _, e := range entries {
		if e.State() != "untranslated" {
			continue
		}
		if e.MsgidPlural != "" {
			e.MsgstrPlural = make([]string, file.Nplurals)
			for i := range e.MsgstrPlural {
				e.MsgstrPlural[i] = "Translated form " + e.Msgid
			}
		} else {
			e.Msgstr = "Translated: " + e.Msgid
		}
	}

	// Write to buffer and re-parse
	var buf bytes.Buffer
	if err := po.Write(&buf, file); err != nil {
		t.Fatalf("write error: %v", err)
	}
	file2, _, err := po.Parse(strings.NewReader(buf.String()))
	if err != nil {
		t.Fatalf("re-parse error: %v", err)
	}

	// Assert all 4 entries translated
	entries2 := activeEntries(file2)
	if len(entries2) != 4 {
		t.Fatalf("expected 4 entries after re-parse, got %d", len(entries2))
	}
	allTranslated := true
	for _, e := range entries2 {
		if e.State() != "translated" {
			allTranslated = false
			t.Errorf("entry %q state = %q", e.Msgid, e.State())
		}
	}
	if !allTranslated {
		t.Error("not all entries are translated after full workflow")
	}
}

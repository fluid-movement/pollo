package po

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

const wrapCol = 76

// Write serialises a File to w in PO format.
func Write(w io.Writer, f *File) error {
	first := true

	writeEntry := func(e *Entry) error {
		var b strings.Builder
		if !first {
			b.WriteByte('\n')
		}
		first = false

		// Translator comment
		if e.TranslatorComment != "" {
			for _, line := range strings.Split(e.TranslatorComment, "\n") {
				b.WriteString("# ")
				b.WriteString(line)
				b.WriteByte('\n')
			}
		}

		// Extracted comment
		if e.ExtractedComment != "" {
			for _, line := range strings.Split(e.ExtractedComment, "\n") {
				b.WriteString("#. ")
				b.WriteString(line)
				b.WriteByte('\n')
			}
		}

		// References
		if len(e.References) > 0 {
			b.WriteString("#: ")
			b.WriteString(strings.Join(e.References, " "))
			b.WriteByte('\n')
		}

		// Flags
		if len(e.Flags) > 0 {
			b.WriteString("#, ")
			b.WriteString(strings.Join(e.Flags, ", "))
			b.WriteByte('\n')
		}

		// Previous msgid
		if e.PrevMsgid != "" {
			b.WriteString(WrapValue("#| msgid", escape(e.PrevMsgid), wrapCol))
		}

		// Previous msgid_plural
		if e.PrevMsgidPlural != "" {
			b.WriteString(WrapValue("#| msgid_plural", escape(e.PrevMsgidPlural), wrapCol))
		}

		// msgctxt
		if e.Msgctxt != "" {
			b.WriteString(WrapValue("msgctxt", escape(e.Msgctxt), wrapCol))
		}

		// msgid
		b.WriteString(WrapValue("msgid", escape(e.Msgid), wrapCol))

		// msgid_plural
		if e.MsgidPlural != "" {
			b.WriteString(WrapValue("msgid_plural", escape(e.MsgidPlural), wrapCol))
		}

		// msgstr (singular) or msgstr[n] (plural)
		if e.MsgidPlural != "" {
			for i, s := range e.MsgstrPlural {
				b.WriteString(WrapValue(fmt.Sprintf("msgstr[%d]", i), escape(s), wrapCol))
			}
		} else {
			b.WriteString(WrapValue("msgstr", escape(e.Msgstr), wrapCol))
		}

		_, err := io.WriteString(w, b.String())
		return err
	}

	// Write header entry first (no preceding blank line)
	if f.Header != nil {
		if err := writeEntry(f.Header); err != nil {
			return err
		}
	}

	// Write remaining nodes in order
	for _, node := range f.Nodes {
		switch n := node.(type) {
		case *Entry:
			if err := writeEntry(n); err != nil {
				return err
			}
		case *ObsoleteNode:
			var b strings.Builder
			if !first {
				b.WriteByte('\n')
			}
			first = false
			for _, line := range n.Lines {
				b.WriteString(line)
				b.WriteByte('\n')
			}
			if _, err := io.WriteString(w, b.String()); err != nil {
				return err
			}
		}
	}

	return nil
}

// escape encodes a decoded string to PO-safe escaped form.
func escape(s string) string {
	var b strings.Builder
	for _, r := range s {
		switch r {
		case '\\':
			b.WriteString(`\\`)
		case '"':
			b.WriteString(`\"`)
		case '\n':
			b.WriteString(`\n`)
		case '\t':
			b.WriteString(`\t`)
		case '\r':
			b.WriteString(`\r`)
		case '\a':
			b.WriteString(`\a`)
		case '\b':
			b.WriteString(`\b`)
		case '\f':
			b.WriteString(`\f`)
		case '\v':
			b.WriteString(`\v`)
		default:
			b.WriteRune(r)
		}
	}
	return b.String()
}

// WriteFile atomically writes a File to the given path.
// It writes to a temp file in the same directory, then renames.
func WriteFile(path string, f *File) error {
	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, ".pollo-*.po.tmp")
	if err != nil {
		return fmt.Errorf("creating temp file: %w", err)
	}
	tmpName := tmp.Name()

	if err := Write(tmp, f); err != nil {
		tmp.Close()
		os.Remove(tmpName)
		return fmt.Errorf("writing temp file: %w", err)
	}

	if err := tmp.Close(); err != nil {
		os.Remove(tmpName)
		return fmt.Errorf("closing temp file: %w", err)
	}

	if err := os.Rename(tmpName, path); err != nil {
		os.Remove(tmpName)
		return fmt.Errorf("renaming temp file: %w", err)
	}

	return nil
}

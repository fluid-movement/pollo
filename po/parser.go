package po

import (
	"bufio"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
)

// Parse parses a PO file from r.
// Returns the parsed File, a slice of warnings, and a fatal error if any.
func Parse(r io.Reader) (*File, []string, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, nil, fmt.Errorf("reading input: %w", err)
	}

	// Normalize line endings: strip \r
	content := strings.ReplaceAll(string(data), "\r\n", "\n")
	content = strings.ReplaceAll(content, "\r", "\n")

	lines := strings.Split(content, "\n")
	// Remove trailing empty line from Split if file ends with \n
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}

	p := &parser{
		lines: lines,
	}

	if err := p.parse(); err != nil {
		return nil, p.warnings, err
	}

	f := &File{
		Nodes:    p.nodes,
		Header:   p.header,
		Language: p.language,
	}

	// Resolve language name
	if f.Language != "" {
		f.LanguageName = LookupLanguageName(f.Language)
	}

	// Resolve nplurals
	hasPluralEntries := false
	for _, node := range f.Nodes {
		if e, ok := node.(*Entry); ok {
			if e.MsgidPlural != "" {
				hasPluralEntries = true
				break
			}
		}
	}

	if p.pluralFormsHeader != "" {
		n, ok := parsePluralForms(p.pluralFormsHeader)
		if ok {
			f.Nplurals = n
		} else if f.Language != "" {
			n2, ok2 := LookupNplurals(f.Language)
			if ok2 {
				f.Nplurals = n2
				p.warnings = append(p.warnings, fmt.Sprintf("Plural-Forms header unparseable; using CLDR fallback nplurals=%d for language %q", n2, f.Language))
			} else if hasPluralEntries {
				return nil, p.warnings, fmt.Errorf("cannot resolve nplurals: Plural-Forms header invalid and language %q not in CLDR table", f.Language)
			} else {
				f.Nplurals = 1
			}
		} else if hasPluralEntries {
			return nil, p.warnings, fmt.Errorf("cannot resolve nplurals: Plural-Forms header invalid and no language set")
		} else {
			f.Nplurals = 1
		}
	} else if f.Language != "" {
		n, ok := LookupNplurals(f.Language)
		if ok {
			f.Nplurals = n
			if hasPluralEntries {
				p.warnings = append(p.warnings, fmt.Sprintf("Plural-Forms header absent; using CLDR fallback nplurals=%d for language %q", n, f.Language))
			}
		} else if hasPluralEntries {
			return nil, p.warnings, fmt.Errorf("cannot resolve nplurals: no Plural-Forms header and language %q not in CLDR table", f.Language)
		} else {
			f.Nplurals = 1
		}
	} else {
		if hasPluralEntries {
			return nil, p.warnings, fmt.Errorf("cannot resolve nplurals: no Plural-Forms header and no Language header")
		}
		f.Nplurals = 1
	}

	return f, p.warnings, nil
}

type parser struct {
	lines             []string
	pos               int
	nodes             []Node
	header            *Entry
	warnings          []string
	language          string
	pluralFormsHeader string
	seenKeys          map[string]bool
}

func (p *parser) parse() error {
	for p.pos < len(p.lines) {
		line := p.lines[p.pos]
		trimmed := strings.TrimSpace(line)

		if trimmed == "" {
			p.pos++
			continue
		}

		if strings.HasPrefix(trimmed, "#~") {
			p.parseObsolete()
			continue
		}

		entry, err := p.parseEntry()
		if err != nil {
			return err
		}
		if entry != nil {
			p.addEntry(entry)
		}
	}
	return nil
}

func (p *parser) addEntry(e *Entry) {
	if e.Msgid == "" && p.header == nil {
		p.header = e
		p.parseHeaderFields(e)
		return
	}
	// Check for duplicate keys (non-header entries only)
	if e.Msgid != "" {
		if p.seenKeys == nil {
			p.seenKeys = make(map[string]bool)
		}
		key := e.Key()
		if p.seenKeys[key] {
			p.warnings = append(p.warnings, fmt.Sprintf("duplicate key %q", key))
		} else {
			p.seenKeys[key] = true
		}
	}
	p.nodes = append(p.nodes, e)
}

func (p *parser) parseHeaderFields(e *Entry) {
	scanner := bufio.NewScanner(strings.NewReader(e.Msgstr))
	for scanner.Scan() {
		line := scanner.Text()
		if after, ok := strings.CutPrefix(line, "Language:"); ok {
			p.language = strings.TrimSpace(after)
		} else if after, ok := strings.CutPrefix(line, "Plural-Forms:"); ok {
			p.pluralFormsHeader = strings.TrimSpace(after)
		}
	}
}

func (p *parser) parseObsolete() {
	var obsLines []string
	for p.pos < len(p.lines) {
		line := p.lines[p.pos]
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "#~") {
			obsLines = append(obsLines, line)
			p.pos++
		} else {
			break
		}
	}
	if len(obsLines) > 0 {
		p.nodes = append(p.nodes, &ObsoleteNode{Lines: obsLines})
	}
}

func (p *parser) parseEntry() (*Entry, error) {
	e := &Entry{}
	var prevMsgidParts []string
	var prevMsgidPluralParts []string
	inPrevMsgid := false
	inPrevMsgidPlural := false

	// Parse comment/flag lines
	for p.pos < len(p.lines) {
		line := p.lines[p.pos]
		trimmed := strings.TrimSpace(line)

		if trimmed == "" {
			break
		}
		if strings.HasPrefix(trimmed, "#~") {
			break
		}
		if isKeywordStart(trimmed) {
			break
		}

		if trimmed == "#" || strings.HasPrefix(trimmed, "# ") {
			comment := ""
			if len(trimmed) > 2 {
				comment = trimmed[2:]
			}
			if e.TranslatorComment == "" {
				e.TranslatorComment = comment
			} else {
				e.TranslatorComment += "\n" + comment
			}
			p.pos++
			inPrevMsgid = false
			inPrevMsgidPlural = false
			continue
		}

		if strings.HasPrefix(trimmed, "#.") {
			comment := trimmed[2:]
			if strings.HasPrefix(comment, " ") {
				comment = comment[1:]
			}
			if e.ExtractedComment == "" {
				e.ExtractedComment = comment
			} else {
				e.ExtractedComment += "\n" + comment
			}
			p.pos++
			inPrevMsgid = false
			inPrevMsgidPlural = false
			continue
		}

		if strings.HasPrefix(trimmed, "#:") {
			refs := strings.TrimSpace(trimmed[2:])
			if refs != "" {
				e.References = append(e.References, strings.Fields(refs)...)
			}
			p.pos++
			inPrevMsgid = false
			inPrevMsgidPlural = false
			continue
		}

		if strings.HasPrefix(trimmed, "#,") {
			flags := strings.TrimSpace(trimmed[2:])
			for _, f := range strings.Split(flags, ",") {
				f = strings.TrimSpace(f)
				if f != "" {
					e.Flags = append(e.Flags, f)
				}
			}
			p.pos++
			inPrevMsgid = false
			inPrevMsgidPlural = false
			continue
		}

		if strings.HasPrefix(trimmed, "#| msgid_plural") {
			rest := strings.TrimSpace(trimmed[len("#| msgid_plural"):])
			val, err := unquote(rest)
			if err != nil {
				return nil, fmt.Errorf("line %d: %w", p.pos+1, err)
			}
			unescaped, warns := unescape(val)
			p.warnings = append(p.warnings, warns...)
			prevMsgidPluralParts = append(prevMsgidPluralParts, unescaped)
			p.pos++
			inPrevMsgid = false
			inPrevMsgidPlural = true
			continue
		}

		if strings.HasPrefix(trimmed, "#| msgid") {
			rest := strings.TrimSpace(trimmed[len("#| msgid"):])
			val, err := unquote(rest)
			if err != nil {
				return nil, fmt.Errorf("line %d: %w", p.pos+1, err)
			}
			unescaped, warns := unescape(val)
			p.warnings = append(p.warnings, warns...)
			prevMsgidParts = append(prevMsgidParts, unescaped)
			p.pos++
			inPrevMsgid = true
			inPrevMsgidPlural = false
			continue
		}

		// Continuation line for #| previous strings
		if strings.HasPrefix(trimmed, "\"") && (inPrevMsgid || inPrevMsgidPlural) {
			val, err := unquote(trimmed)
			if err != nil {
				return nil, fmt.Errorf("line %d: %w", p.pos+1, err)
			}
			unescaped, warns := unescape(val)
			p.warnings = append(p.warnings, warns...)
			if inPrevMsgid && len(prevMsgidParts) > 0 {
				prevMsgidParts[len(prevMsgidParts)-1] += unescaped
			} else if inPrevMsgidPlural && len(prevMsgidPluralParts) > 0 {
				prevMsgidPluralParts[len(prevMsgidPluralParts)-1] += unescaped
			}
			p.pos++
			continue
		}

		// Unknown comment type - skip
		if strings.HasPrefix(trimmed, "#") {
			p.pos++
			inPrevMsgid = false
			inPrevMsgidPlural = false
			continue
		}

		break
	}

	if len(prevMsgidParts) > 0 {
		e.PrevMsgid = strings.Join(prevMsgidParts, "")
	}
	if len(prevMsgidPluralParts) > 0 {
		e.PrevMsgidPlural = strings.Join(prevMsgidPluralParts, "")
	}

	if p.pos >= len(p.lines) {
		return nil, nil
	}

	line := p.lines[p.pos]
	trimmed := strings.TrimSpace(line)

	if trimmed == "" || strings.HasPrefix(trimmed, "#~") {
		return nil, nil
	}

	// msgctxt
	if strings.HasPrefix(trimmed, "msgctxt ") || strings.HasPrefix(trimmed, "msgctxt\t") {
		val, err := p.parseKeywordValue("msgctxt")
		if err != nil {
			return nil, err
		}
		e.Msgctxt = val
	}

	if p.pos >= len(p.lines) {
		if e.Msgctxt != "" {
			return nil, fmt.Errorf("unexpected end of file after msgctxt")
		}
		return nil, nil
	}

	line = p.lines[p.pos]
	trimmed = strings.TrimSpace(line)

	if !strings.HasPrefix(trimmed, "msgid ") && !strings.HasPrefix(trimmed, "msgid\t") && trimmed != `msgid ""` {
		if e.Msgctxt != "" {
			return nil, fmt.Errorf("line %d: expected msgid after msgctxt, got %q", p.pos+1, trimmed)
		}
		return nil, nil
	}

	val, err := p.parseKeywordValue("msgid")
	if err != nil {
		return nil, err
	}
	e.Msgid = val

	// Check for Content-Type charset in header
	// (We need to check this after parsing the full entry msgstr)

	if p.pos >= len(p.lines) {
		return nil, fmt.Errorf("unexpected end of file after msgid")
	}

	line = p.lines[p.pos]
	trimmed = strings.TrimSpace(line)

	// msgid_plural
	if strings.HasPrefix(trimmed, "msgid_plural ") || strings.HasPrefix(trimmed, "msgid_plural\t") {
		val, err = p.parseKeywordValue("msgid_plural")
		if err != nil {
			return nil, err
		}
		e.MsgidPlural = val
	}

	if p.pos >= len(p.lines) {
		return nil, fmt.Errorf("unexpected end of file after msgid")
	}

	line = p.lines[p.pos]
	trimmed = strings.TrimSpace(line)

	if e.MsgidPlural != "" {
		// Parse msgstr[n]
		pluralForms := map[int]string{}
		for p.pos < len(p.lines) {
			line = p.lines[p.pos]
			trimmed = strings.TrimSpace(line)
			if trimmed == "" || strings.HasPrefix(trimmed, "#") || strings.HasPrefix(trimmed, "#~") {
				break
			}
			m := msgstrPluralRe.FindStringSubmatch(trimmed)
			if m == nil {
				break
			}
			idx, _ := strconv.Atoi(m[1])
			rest := strings.TrimSpace(trimmed[len(m[0]):])
			val, err := p.parseInlineAndContinuations(rest)
			if err != nil {
				return nil, fmt.Errorf("line %d: %w", p.pos, err)
			}
			pluralForms[idx] = val
		}
		if len(pluralForms) > 0 {
			for i := 0; i < len(pluralForms); i++ {
				if _, ok := pluralForms[i]; !ok {
					return nil, fmt.Errorf("non-contiguous msgstr[n] indices: missing index %d", i)
				}
			}
			e.MsgstrPlural = make([]string, len(pluralForms))
			for i := range len(pluralForms) {
				e.MsgstrPlural[i] = pluralForms[i]
			}
		}
	} else {
		if strings.HasPrefix(trimmed, "msgstr") {
			val, err = p.parseKeywordValue("msgstr")
			if err != nil {
				return nil, err
			}
			e.Msgstr = val
		} else if trimmed != "" && !strings.HasPrefix(trimmed, "#") && !strings.HasPrefix(trimmed, "#~") {
			return nil, fmt.Errorf("line %d: expected msgstr, got %q", p.pos+1, trimmed)
		}
	}

	// Check charset for header entry (empty msgid)
	if e.Msgid == "" {
		// Check Content-Type in the msgstr
		scanner := bufio.NewScanner(strings.NewReader(e.Msgstr))
		for scanner.Scan() {
			hline := scanner.Text()
			if after, ok := strings.CutPrefix(hline, "Content-Type:"); ok {
				if err := checkCharset(after); err != nil {
					return nil, err
				}
			}
		}
	}

	return e, nil
}

var msgstrPluralRe = regexp.MustCompile(`^msgstr\[(\d+)\]\s*`)

func isKeywordStart(s string) bool {
	for _, kw := range []string{"msgid ", "msgid\t", `msgid ""`, "msgstr", "msgctxt ", "msgctxt\t", "msgid_plural "} {
		if strings.HasPrefix(s, kw) {
			return true
		}
	}
	return false
}

// parseKeywordValue parses a keyword line and its continuation lines.
func (p *parser) parseKeywordValue(keyword string) (string, error) {
	line := p.lines[p.pos]
	trimmed := strings.TrimSpace(line)

	rest := trimmed[len(keyword):]
	rest = strings.TrimLeft(rest, " \t")

	return p.parseInlineAndContinuations(rest)
}

// parseInlineAndContinuations parses a quoted string and any following continuation lines.
func (p *parser) parseInlineAndContinuations(rest string) (string, error) {
	p.pos++ // consume current line

	firstQuoted, err := unquote(rest)
	if err != nil {
		return "", err
	}
	unescaped, warns := unescape(firstQuoted)
	p.warnings = append(p.warnings, warns...)
	result := unescaped

	// Continuation lines
	for p.pos < len(p.lines) {
		line := p.lines[p.pos]
		trimmed := strings.TrimSpace(line)
		if !strings.HasPrefix(trimmed, "\"") {
			break
		}
		val, err := unquote(trimmed)
		if err != nil {
			return "", fmt.Errorf("line %d: %w", p.pos+1, err)
		}
		unescaped, warns := unescape(val)
		p.warnings = append(p.warnings, warns...)
		result += unescaped
		p.pos++
	}

	return result, nil
}

// unquote strips surrounding double quotes from a PO string literal.
func unquote(s string) (string, error) {
	s = strings.TrimSpace(s)
	if len(s) < 2 || s[0] != '"' || s[len(s)-1] != '"' {
		return "", fmt.Errorf("expected quoted string, got %q", s)
	}
	return s[1 : len(s)-1], nil
}

// unescape decodes PO escape sequences. Unknown sequences pass through with a warning.
func unescape(s string) (string, []string) {
	var b strings.Builder
	var warns []string
	i := 0
	for i < len(s) {
		if s[i] != '\\' {
			b.WriteByte(s[i])
			i++
			continue
		}
		i++
		if i >= len(s) {
			b.WriteByte('\\')
			break
		}
		switch s[i] {
		case 'n':
			b.WriteByte('\n')
		case 't':
			b.WriteByte('\t')
		case 'r':
			b.WriteByte('\r')
		case '\\':
			b.WriteByte('\\')
		case '"':
			b.WriteByte('"')
		case 'a':
			b.WriteByte('\a')
		case 'b':
			b.WriteByte('\b')
		case 'f':
			b.WriteByte('\f')
		case 'v':
			b.WriteByte('\v')
		default:
			b.WriteByte('\\')
			b.WriteByte(s[i])
			warns = append(warns, fmt.Sprintf("unrecognised escape sequence \\%c", s[i]))
		}
		i++
	}
	return b.String(), warns
}

var charsetRe = regexp.MustCompile(`(?i)charset\s*=\s*([^\s;,]+)`)

func checkCharset(contentType string) error {
	m := charsetRe.FindStringSubmatch(contentType)
	if m == nil {
		return nil
	}
	charset := m[1]
	if !strings.EqualFold(charset, "utf-8") && !strings.EqualFold(charset, "utf8") {
		return fmt.Errorf("unsupported charset %q: only UTF-8 is supported", charset)
	}
	return nil
}

var npluralsRe = regexp.MustCompile(`nplurals\s*=\s*(\d+)`)

func parsePluralForms(s string) (int, bool) {
	m := npluralsRe.FindStringSubmatch(s)
	if m == nil {
		return 0, false
	}
	n, err := strconv.Atoi(m[1])
	if err != nil || n < 1 {
		return 0, false
	}
	return n, true
}

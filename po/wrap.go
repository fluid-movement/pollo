package po

import (
	"strings"
	"unicode/utf8"
)

// WrapValue formats a PO keyword + value string with line wrapping.
// col is the column limit in runes (0 = disabled, no wrapping).
// The value must already be PO-escaped (e.g. \n, \t, etc.).
// Returns the complete formatted output including the trailing newline.
func WrapValue(keyword, value string, col int) string {
	if value == "" {
		return keyword + " \"\"\n"
	}

	// If wrapping is disabled, always single line.
	if col == 0 {
		return keyword + " \"" + value + "\"\n"
	}

	// Check if value contains \n escape sequences (literal two-char sequences).
	hasNewlineEscape := strings.Contains(value, `\n`)

	if !hasNewlineEscape {
		// Try single-line: keyword + space + quote + value + quote
		single := keyword + " \"" + value + "\""
		if utf8.RuneCountInString(single) <= col {
			return single + "\n"
		}
		// Over limit: go multi-line
		return multiLine(keyword, value, col)
	}

	// Has \n sequences: always multi-line
	return multiLine(keyword, value, col)
}

// multiLine formats the value as a multi-line block:
// keyword ""
// "segment1\n"
// "segment2"
func multiLine(keyword, value string, col int) string {
	var b strings.Builder
	b.WriteString(keyword)
	b.WriteString(" \"\"\n")

	// Split on \n escape sequences, keeping \n at end of each segment.
	segments := splitOnNewlineEscapes(value)

	for _, seg := range segments {
		// Further split long segments.
		parts := splitLongSegment(seg, col)
		for _, p := range parts {
			b.WriteByte('"')
			b.WriteString(p)
			b.WriteString("\"\n")
		}
	}
	return b.String()
}

// splitOnNewlineEscapes splits value at \n escape sequences, keeping \n at end.
// e.g. "foo\nbar\nbaz" → ["foo\n", "bar\n", "baz"]
func splitOnNewlineEscapes(value string) []string {
	var segments []string
	for {
		idx := strings.Index(value, `\n`)
		if idx < 0 {
			segments = append(segments, value)
			break
		}
		segments = append(segments, value[:idx+2]) // include \n
		value = value[idx+2:]
	}
	return segments
}

// splitLongSegment splits a segment at space boundaries to fit within col runes.
// The col limit applies to the quoted line: quote + content + quote = col.
// So content must fit within col-2 runes.
func splitLongSegment(seg string, col int) []string {
	if col == 0 {
		return []string{seg}
	}
	limit := col - 2 // for the surrounding quotes
	if limit <= 0 {
		return []string{seg}
	}
	if utf8.RuneCountInString(seg) <= limit {
		return []string{seg}
	}

	var parts []string
	remaining := seg
	for utf8.RuneCountInString(remaining) > limit {
		// Find last space within limit runes
		cut := lastSpaceWithin(remaining, limit)
		if cut <= 0 {
			// No space found: emit oversized line
			// But we still need to check if there's more after some point.
			// If no space at all in remaining, emit whole thing.
			if !strings.Contains(remaining, " ") {
				parts = append(parts, remaining)
				remaining = ""
				break
			}
			// There's a space, but not within limit. Find first space.
			idx := strings.IndexByte(remaining, ' ')
			if idx < 0 {
				parts = append(parts, remaining)
				remaining = ""
				break
			}
			// Emit up to and including the space? No - emit up to first space.
			// Actually for "no space boundary in segment → emit oversized line"
			// means if no space within limit, emit from start until next space boundary.
			parts = append(parts, remaining[:idx+1])
			remaining = remaining[idx+1:]
		} else {
			parts = append(parts, remaining[:cut+1])
			remaining = remaining[cut+1:]
		}
	}
	if remaining != "" {
		parts = append(parts, remaining)
	}
	return parts
}

// lastSpaceWithin returns the byte index of the last space within the first
// limit runes of s, or -1 if none found.
func lastSpaceWithin(s string, limit int) int {
	lastSpace := -1
	i := 0
	runeCount := 0
	for i < len(s) && runeCount < limit {
		r, size := utf8.DecodeRuneInString(s[i:])
		if r == ' ' {
			lastSpace = i
		}
		i += size
		runeCount++
	}
	return lastSpace
}

package po

import (
	"bufio"
	"errors"
	"os"
	"strings"
)

// ReadSkipFile reads a skip file and returns the set of keys to skip.
// If the file does not exist, returns an empty map with no error.
func ReadSkipFile(path string) (map[string]struct{}, error) {
	f, err := os.Open(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return map[string]struct{}{}, nil
		}
		return nil, err
	}
	defer f.Close()

	result := map[string]struct{}{}
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		// Strip \r\n
		line = strings.TrimRight(line, "\r")

		// Blank lines and comment lines
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Split on first tab: key[\treason]
		key, _, _ := strings.Cut(line, "\t")

		// Unescape the key
		key = unescapeSkipKey(key)

		result[key] = struct{}{}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

// unescapeSkipKey decodes skip file escape sequences in a key.
// \n → newline, \r → CR, \\ → backslash. U+0004 preserved as-is.
func unescapeSkipKey(s string) string {
	var b strings.Builder
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
		case 'r':
			b.WriteByte('\r')
		case '\\':
			b.WriteByte('\\')
		default:
			b.WriteByte('\\')
			b.WriteByte(s[i])
		}
		i++
	}
	return b.String()
}

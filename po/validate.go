package po

import (
	"fmt"
	"regexp"
	"slices"
	"strings"
)

// ValidatePlaceholders checks that placeholders in translations match the source.
// Returns a warning string if there is a mismatch, or "" if all is well.
// Only validates if the entry flags contain a known format marker.
func ValidatePlaceholders(entry *Entry, translations []string) string {
	format := detectFormat(entry.Flags)
	if format == "" {
		return ""
	}

	extract := extractorFor(format)
	if extract == nil {
		return ""
	}

	source := entry.Msgid
	srcPlaceholders := extract(source)

	// Combine all translation forms for validation
	var dstPlaceholders []string
	for _, t := range translations {
		dstPlaceholders = append(dstPlaceholders, extract(t)...)
	}

	// For plural entries, we check each translation form against the source.
	// Actually the spec says multiset equality between source and translation.
	// For plural entries, compare source against each form individually.
	if len(translations) > 1 {
		for _, t := range translations {
			tPlaceholders := extract(t)
			if w := comparePlaceholders(srcPlaceholders, tPlaceholders); w != "" {
				return w
			}
		}
		return ""
	}

	// Singular: compare source against the single translation
	return comparePlaceholders(srcPlaceholders, dstPlaceholders)
}

func detectFormat(flags []string) string {
	for _, f := range flags {
		switch f {
		case "c-format", "objc-format", "javascript-format", "php-format",
			"python-format", "python-brace-format", "sh-format":
			return f
		}
	}
	return ""
}

func extractorFor(format string) func(string) []string {
	switch format {
	case "c-format", "objc-format", "javascript-format", "php-format":
		return extractCFormat
	case "python-format":
		return extractPythonFormat
	case "python-brace-format":
		return extractPythonBraceFormat
	case "sh-format":
		return extractShFormat
	}
	return nil
}

// cFormatRe matches printf-style placeholders: %[flags][width][.prec][len]spec
// and positional %1$s forms. Does NOT match %%.
var cFormatRe = regexp.MustCompile(`%(?:%|(?:\d+\$)?[-+ #0]*(?:\*|\d+)?(?:\.(?:\*|\d+))?(?:hh|ll|[hlLqjzt])?[diouxXeEfgGaAcspn])`)

func extractCFormat(s string) []string {
	matches := cFormatRe.FindAllString(s, -1)
	var result []string
	for _, m := range matches {
		if m == "%%" {
			continue
		}
		result = append(result, m)
	}
	return result
}

// pythonFormatRe matches printf-style + %(name)s named groups.
var pythonFormatRe = regexp.MustCompile(`%(?:%|\([^)]+\)[diouxXeEfgGaAcsp]|(?:\d+\$)?[-+ #0]*(?:\*|\d+)?(?:\.(?:\*|\d+))?(?:hh|ll|[hlLqjzt])?[diouxXeEfgGaAcspn])`)

func extractPythonFormat(s string) []string {
	matches := pythonFormatRe.FindAllString(s, -1)
	var result []string
	for _, m := range matches {
		if m == "%%" {
			continue
		}
		result = append(result, m)
	}
	return result
}

// pythonBraceFormatRe matches {0}, {name}, {value:.2f}, {}, etc.
var pythonBraceFormatRe = regexp.MustCompile(`\{[^{}]*\}`)

func extractPythonBraceFormat(s string) []string {
	return pythonBraceFormatRe.FindAllString(s, -1)
}

// shFormatRe matches $VAR and ${VAR}.
var shFormatRe = regexp.MustCompile(`\$(?:\{[A-Za-z_][A-Za-z0-9_]*\}|[A-Za-z_][A-Za-z0-9_]*)`)

func extractShFormat(s string) []string {
	return shFormatRe.FindAllString(s, -1)
}

func comparePlaceholders(src, dst []string) string {
	if slicesEqual(src, dst) {
		return ""
	}
	srcSorted := sortedCopy(src)
	dstSorted := sortedCopy(dst)
	return fmt.Sprintf("placeholder mismatch: source has [%s] but translation has [%s]",
		strings.Join(srcSorted, ", "),
		strings.Join(dstSorted, ", "))
}

func slicesEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	ac := sortedCopy(a)
	bc := sortedCopy(b)
	for i := range ac {
		if ac[i] != bc[i] {
			return false
		}
	}
	return true
}

func sortedCopy(s []string) []string {
	c := make([]string, len(s))
	copy(c, s)
	slices.Sort(c)
	return c
}

package po

import "strings"

// Entry represents a single active PO entry (translation unit).
type Entry struct {
	TranslatorComment  string
	ExtractedComment   string
	References         []string
	Flags              []string
	PrevMsgid          string
	PrevMsgidPlural    string
	Msgctxt            string
	Msgid              string
	MsgidPlural        string
	Msgstr             string   // singular
	MsgstrPlural       []string // plural forms [0..n-1]
}

// Key returns the unique lookup key for this entry.
// Format: msgctxt + U+0004 + msgid if context present, else just msgid.
func (e *Entry) Key() string {
	if e.Msgctxt != "" {
		return e.Msgctxt + "\x04" + e.Msgid
	}
	return e.Msgid
}

// State returns "translated", "fuzzy", or "untranslated".
func (e *Entry) State() string {
	for _, f := range e.Flags {
		if f == "fuzzy" {
			return "fuzzy"
		}
	}
	if e.MsgidPlural != "" {
		for _, s := range e.MsgstrPlural {
			if isBlank(s) {
				return "untranslated"
			}
		}
		return "translated"
	}
	if isBlank(e.Msgstr) {
		return "untranslated"
	}
	return "translated"
}

func isBlank(s string) bool {
	return strings.TrimSpace(s) == ""
}

// ObsoleteNode holds a contiguous block of #~ lines verbatim.
type ObsoleteNode struct {
	Lines []string
}

// Node is either an *Entry or an *ObsoleteNode.
type Node interface {
	isNode()
}

func (*Entry) isNode()        {}
func (*ObsoleteNode) isNode() {}

// File is a parsed PO file.
type File struct {
	Nodes       []Node
	Header      *Entry
	Nplurals    int
	Language    string
	LanguageName string
}

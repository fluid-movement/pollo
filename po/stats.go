package po

// CountEntries counts the state of all active entries, excluding the header and obsolete nodes.
func CountEntries(f *File) (total, translated, fuzzy, untranslated int) {
	for _, node := range f.Nodes {
		e, ok := node.(*Entry)
		if !ok {
			continue
		}
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
	return
}

// ComputeRemaining counts fuzzy and untranslated entries, excluding any keys present in skipMap.
func ComputeRemaining(f *File, skipMap map[string]struct{}) (remaining, fuzzy, untranslated int) {
	for _, node := range f.Nodes {
		e, ok := node.(*Entry)
		if !ok {
			continue
		}
		if _, skip := skipMap[e.Key()]; skip {
			continue
		}
		switch e.State() {
		case "fuzzy":
			fuzzy++
			remaining++
		case "untranslated":
			untranslated++
			remaining++
		}
	}
	return
}

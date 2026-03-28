package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/yourusername/pollo/po"
)

var statsCmd = &cobra.Command{
	Use:   "stats <file>",
	Short: "Report translation progress for a PO file",
	Args:  cobra.ExactArgs(1),
	Run:   runStats,
}

func runStats(cmd *cobra.Command, args []string) {
	path := args[0]

	f, err := os.Open(path)
	if err != nil {
		writeError(fmt.Sprintf("opening file: %s", err))
	}
	defer f.Close()

	file, warnings, err := po.Parse(f)
	if err != nil {
		writeError(fmt.Sprintf("parse error: %s", err))
	}

	total, translated, fuzzy, untranslated := countEntries(file)

	warningsArr := warnings
	if warningsArr == nil {
		warningsArr = []string{}
	}

	_ = writeJSON(map[string]any{
		"file":           path,
		"language":       file.Language,
		"language_name":  file.LanguageName,
		"total":          total,
		"translated":     translated,
		"fuzzy":          fuzzy,
		"untranslated":   untranslated,
		"remaining":      fuzzy + untranslated,
		"parse_warnings": warningsArr,
	})
}

// countEntries counts translation entry states, excluding header and obsolete.
func countEntries(file *po.File) (total, translated, fuzzy, untranslated int) {
	for _, node := range file.Nodes {
		e, ok := node.(*po.Entry)
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

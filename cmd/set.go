package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/yourusername/pollo/po"
)

var setCmd = &cobra.Command{
	Use:   "set <file>",
	Short: "Write a confirmed translation to a PO file",
	Args:  cobra.ExactArgs(1),
	Run:   runSet,
}

var (
	setIDFlag               string
	setIDFileFlag           string
	setContextFlag          string
	setContextFileFlag      string
	setTranslationFlag      string
	setTranslationFileFlag  string
	setTranslationsFlag     string
	setTranslationsFileFlag string
)

func init() {
	setCmd.Flags().StringVar(&setIDFlag, "id", "", "msgid of the target entry")
	setCmd.Flags().StringVar(&setIDFileFlag, "id-file", "", "Read msgid from file")
	setCmd.Flags().StringVar(&setContextFlag, "context", "", "msgctxt of the target entry")
	setCmd.Flags().StringVar(&setContextFileFlag, "context-file", "", "Read msgctxt from file")
	setCmd.Flags().StringVar(&setTranslationFlag, "translation", "", "Translated string (singular)")
	setCmd.Flags().StringVar(&setTranslationFileFlag, "translation-file", "", "Read translation from file")
	setCmd.Flags().StringVar(&setTranslationsFlag, "translations", "", "JSON array of plural forms")
	setCmd.Flags().StringVar(&setTranslationsFileFlag, "translations-file", "", "Read JSON array from file")
}

func runSet(cmd *cobra.Command, args []string) {
	path := args[0]

	// Validate --id / --id-file: exactly one required
	hasID := setIDFlag != "" || setIDFileFlag != ""
	if !hasID {
		writeError("exactly one of --id or --id-file is required")
	}
	if setIDFlag != "" && setIDFileFlag != "" {
		writeError("--id and --id-file are mutually exclusive")
	}

	// Validate translation flags: exactly one required
	tFlagsChanged := 0
	if cmd.Flags().Changed("translation") {
		tFlagsChanged++
	}
	if cmd.Flags().Changed("translation-file") {
		tFlagsChanged++
	}
	if cmd.Flags().Changed("translations") {
		tFlagsChanged++
	}
	if cmd.Flags().Changed("translations-file") {
		tFlagsChanged++
	}

	if tFlagsChanged == 0 {
		writeError("exactly one of --translation, --translation-file, --translations, or --translations-file is required")
	}
	if tFlagsChanged > 1 {
		writeError("only one of --translation, --translation-file, --translations, or --translations-file may be specified")
	}

	// Read ID
	msgid, err := readFlagOrFile(setIDFlag, setIDFileFlag, "id")
	if err != nil {
		writeError(err.Error())
	}

	// Read context (optional)
	ctx := ""
	if cmd.Flags().Changed("context") || cmd.Flags().Changed("context-file") {
		ctx, err = readFlagOrFile(setContextFlag, setContextFileFlag, "context")
		if err != nil {
			writeError(err.Error())
		}
	}

	// Build key
	key := msgid
	if ctx != "" {
		key = ctx + "\x04" + msgid
	}

	// Parse file
	fh, err := os.Open(path)
	if err != nil {
		writeError(fmt.Sprintf("opening file: %s", err))
	}

	file, parseWarnings, err := po.Parse(fh)
	fh.Close()
	if err != nil {
		writeError(fmt.Sprintf("parse error: %s", err))
	}

	warningsArr := parseWarnings
	if warningsArr == nil {
		warningsArr = []string{}
	}

	// Find entry
	var entry *po.Entry
	for _, node := range file.Nodes {
		e, ok := node.(*po.Entry)
		if !ok {
			continue
		}
		if e.Key() == key {
			entry = e
			break
		}
	}

	if entry == nil {
		_ = writeJSON(map[string]any{
			"ok":    false,
			"error": fmt.Sprintf("entry not found: %q", msgid),
		})
		os.Exit(1)
	}

	isPlural := entry.MsgidPlural != ""

	// Determine which translation flag was used
	isPluralsFlag := cmd.Flags().Changed("translations") || cmd.Flags().Changed("translations-file")
	isSingularFlag := cmd.Flags().Changed("translation") || cmd.Flags().Changed("translation-file")

	// Validate singular/plural mismatch
	if isPlural && isSingularFlag {
		_ = writeJSON(map[string]any{
			"ok":    false,
			"error": "plural entry requires --translations or --translations-file",
		})
		os.Exit(1)
	}
	if !isPlural && isPluralsFlag {
		_ = writeJSON(map[string]any{
			"ok":    false,
			"error": "singular entry does not accept --translations or --translations-file",
		})
		os.Exit(1)
	}

	var warning string

	if isPlural {
		// Read translations JSON array
		rawJSON, err := readFlagOrFile(setTranslationsFlag, setTranslationsFileFlag, "translations")
		if err != nil {
			writeError(err.Error())
		}

		var forms []string
		if err := json.Unmarshal([]byte(rawJSON), &forms); err != nil {
			writeError(fmt.Sprintf("parsing --translations JSON: %s", err))
		}

		if len(forms) != file.Nplurals {
			_ = writeJSON(map[string]any{
				"ok":    false,
				"error": fmt.Sprintf("plural array length %d does not match nplurals=%d", len(forms), file.Nplurals),
			})
			os.Exit(1)
		}

		// Validate placeholders
		warning = po.ValidatePlaceholders(entry, forms)

		// Apply translation
		entry.MsgstrPlural = forms
		entry.Flags = removeFuzzy(entry.Flags)
		entry.PrevMsgid = ""
		entry.PrevMsgidPlural = ""

	} else {
		// Singular
		translation, err := readFlagOrFile(setTranslationFlag, setTranslationFileFlag, "translation")
		if err != nil {
			writeError(err.Error())
		}

		// Validate placeholders
		warning = po.ValidatePlaceholders(entry, []string{translation})

		// Apply translation
		entry.Msgstr = translation
		entry.Flags = removeFuzzy(entry.Flags)
		entry.PrevMsgid = ""
		entry.PrevMsgidPlural = ""
	}

	// Write file atomically
	if err := po.WriteFile(path, file); err != nil {
		writeError(fmt.Sprintf("writing file: %s", err))
	}

	// Compute remaining from in-memory state
	_, _, fuzzy, untranslated := po.CountEntries(file)
	remaining := fuzzy + untranslated

	resp := map[string]any{
		"ok":                      true,
		"file":                    path,
		"msgid":                   msgid,
		"remaining":               remaining,
		"remaining_fuzzy":         fuzzy,
		"remaining_untranslated":  untranslated,
		"parse_warnings":          warningsArr,
	}
	if warning != "" {
		resp["warning"] = warning
	}

	_ = writeJSON(resp)
}

func removeFuzzy(flags []string) []string {
	var result []string
	for _, f := range flags {
		if f != "fuzzy" {
			result = append(result, f)
		}
	}
	return result
}

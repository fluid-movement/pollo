package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/yourusername/pollo/po"
)

var getCmd = &cobra.Command{
	Use:   "get <file>",
	Short: "Return the next entry needing translation",
	Args:  cobra.ExactArgs(1),
	Run:   runGet,
}

var (
	getIDFlag              string
	getIDFileFlag          string
	getContextFlag         string
	getContextFileFlag     string
	getSkipIDsFileFlag     string
	getOrderFlag           string
	getIncludeTranslated   bool
)

func init() {
	getCmd.Flags().StringVar(&getIDFlag, "id", "", "Fetch entry by msgid")
	getCmd.Flags().StringVar(&getIDFileFlag, "id-file", "", "Read msgid from file")
	getCmd.Flags().StringVar(&getContextFlag, "context", "", "msgctxt for the entry")
	getCmd.Flags().StringVar(&getContextFileFlag, "context-file", "", "Read msgctxt from file")
	getCmd.Flags().StringVar(&getSkipIDsFileFlag, "skip-ids-file", "", "Path to skip file")
	getCmd.Flags().StringVar(&getOrderFlag, "order", "fuzzy-first", "Iteration order: fuzzy-first or untranslated-first")
	getCmd.Flags().BoolVar(&getIncludeTranslated, "include-translated", false, "Return all entries regardless of state")
}

func runGet(cmd *cobra.Command, args []string) {
	path := args[0]

	// Validate mutual exclusions
	byID := getIDFlag != "" || getIDFileFlag != ""
	if byID && (cmd.Flags().Changed("order") || cmd.Flags().Changed("skip-ids-file") || cmd.Flags().Changed("include-translated")) {
		writeError("--id/--id-file are mutually exclusive with --order, --skip-ids-file, and --include-translated")
	}

	f, err := os.Open(path)
	if err != nil {
		writeError(fmt.Sprintf("opening file: %s", err))
	}
	defer f.Close()

	file, warnings, err := po.Parse(f)
	if err != nil {
		writeError(fmt.Sprintf("parse error: %s", err))
	}

	warningsArr := warnings
	if warningsArr == nil {
		warningsArr = []string{}
	}

	if byID {
		runGetByID(path, file, warningsArr)
		return
	}

	// Load skip file
	skipMap := map[string]struct{}{}
	if getSkipIDsFileFlag != "" {
		skipMap, err = po.ReadSkipFile(getSkipIDsFileFlag)
		if err != nil {
			writeError(fmt.Sprintf("reading skip file: %s", err))
		}
	}

	runGetIterate(path, file, warningsArr, skipMap)
}

func runGetByID(path string, file *po.File, warnings []string) {
	msgid, err := readFlagOrFile(getIDFlag, getIDFileFlag, "id")
	if err != nil {
		writeError(err.Error())
	}

	ctx := ""
	if getContextFlag != "" || getContextFileFlag != "" {
		ctx, err = readFlagOrFile(getContextFlag, getContextFileFlag, "context")
		if err != nil {
			writeError(err.Error())
		}
	}

	key := msgid
	if ctx != "" {
		key = ctx + "\x04" + msgid
	}

	var found *po.Entry
	for _, node := range file.Nodes {
		e, ok := node.(*po.Entry)
		if !ok {
			continue
		}
		if e.Key() == key {
			found = e
			break
		}
	}

	if found == nil {
		_ = writeJSON(map[string]any{
			"ok":    false,
			"error": fmt.Sprintf("entry not found: %q", msgid),
		})
		os.Exit(2)
	}

	_, _, fuzzy, untranslated := po.CountEntries(file)
	_ = writeJSON(buildEntryResponse(false, path, file, found, fuzzy+untranslated, fuzzy, untranslated, warnings))
}

func runGetIterate(path string, file *po.File, warnings []string, skipMap map[string]struct{}) {
	var candidates []*po.Entry

	if getIncludeTranslated {
		for _, node := range file.Nodes {
			e, ok := node.(*po.Entry)
			if !ok {
				continue
			}
			if _, skip := skipMap[e.Key()]; skip {
				continue
			}
			candidates = append(candidates, e)
		}
	} else {
		// Collect by order
		var fuzzyEntries, untranslatedEntries []*po.Entry
		for _, node := range file.Nodes {
			e, ok := node.(*po.Entry)
			if !ok {
				continue
			}
			if _, skip := skipMap[e.Key()]; skip {
				continue
			}
			switch e.State() {
			case "fuzzy":
				fuzzyEntries = append(fuzzyEntries, e)
			case "untranslated":
				untranslatedEntries = append(untranslatedEntries, e)
			}
		}

		if getOrderFlag == "untranslated-first" {
			candidates = append(untranslatedEntries, fuzzyEntries...)
		} else {
			// fuzzy-first (default)
			candidates = append(fuzzyEntries, untranslatedEntries...)
		}
	}

	// Compute remaining counts (skip-file respected)
	remaining, remainingFuzzy, remainingUntranslated := po.ComputeRemaining(file, skipMap)

	if len(candidates) == 0 {
		_ = writeJSON(map[string]any{
			"done":                    true,
			"file":                    path,
			"language":                file.Language,
			"language_name":           file.LanguageName,
			"remaining":               remaining,
			"remaining_fuzzy":         remainingFuzzy,
			"remaining_untranslated":  remainingUntranslated,
		})
		return
	}

	entry := candidates[0]
	_ = writeJSON(buildEntryResponse(false, path, file, entry, remaining, remainingFuzzy, remainingUntranslated, warnings))
}


func buildEntryResponse(done bool, path string, file *po.File, e *po.Entry, remaining, remainingFuzzy, remainingUntranslated int, warnings []string) map[string]any {
	// Build flags without "fuzzy"
	var flags []string
	for _, f := range e.Flags {
		if f != "fuzzy" {
			flags = append(flags, f)
		}
	}
	if flags == nil {
		flags = []string{}
	}

	resp := map[string]any{
		"done":          done,
		"file":          path,
		"language":      file.Language,
		"language_name": file.LanguageName,
		"msgid":         e.Msgid,
		"state":         e.State(),
		"flags":         flags,
		"parse_warnings": warnings,
		"remaining":               remaining,
		"remaining_fuzzy":         remainingFuzzy,
		"remaining_untranslated":  remainingUntranslated,
	}

	// Optional fields always present as null or value
	setNullableString(resp, "msgid_plural", e.MsgidPlural)
	setNullableString(resp, "msgctxt", e.Msgctxt)
	setNullableString(resp, "translator_comment", e.TranslatorComment)
	setNullableString(resp, "extracted_comment", e.ExtractedComment)
	setNullableString(resp, "previous_msgid", e.PrevMsgid)
	setNullableString(resp, "previous_msgid_plural", e.PrevMsgidPlural)

	// current_msgstr vs current_msgstr_plural (mutually null)
	if e.MsgidPlural != "" {
		resp["current_msgstr"] = nil
		if e.MsgstrPlural != nil {
			resp["current_msgstr_plural"] = e.MsgstrPlural
		} else {
			resp["current_msgstr_plural"] = []string{}
		}
		resp["plural_count"] = file.Nplurals
	} else {
		resp["current_msgstr"] = e.Msgstr
		resp["current_msgstr_plural"] = nil
		resp["plural_count"] = nil
	}

	return resp
}

func setNullableString(m map[string]any, key, val string) {
	if val == "" {
		m[key] = nil
	} else {
		m[key] = val
	}
}

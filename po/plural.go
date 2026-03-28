package po

import "strings"

// npluralsTable maps BCP 47 language tags (lowercased) to nplurals counts.
var npluralsTable = map[string]int{
	// 1 plural form
	"ja": 1, "zh": 1, "ko": 1, "tr": 1, "vi": 1, "th": 1, "id": 1,
	"ms": 1, "fa": 1, "my": 1, "km": 1, "lo": 1, "ka": 1, "az": 1, "kn": 1,
	"zh-hans": 1, "zh-hant": 1, "zh-tw": 1, "zh-hk": 1,

	// 2 plural forms
	"en": 2, "de": 2, "nl": 2, "sv": 2, "da": 2, "nb": 2, "nn": 2, "fi": 2,
	"et": 2, "hu": 2, "el": 2, "he": 2, "it": 2, "es": 2, "pt": 2, "ca": 2,
	"af": 2, "bg": 2, "sq": 2, "hy": 2, "eu": 2, "gl": 2, "ur": 2, "hi": 2,
	"bn": 2, "fr": 2, "pt-br": 2, "de-at": 2, "de-ch": 2, "ro": 2,
	"en-gb": 2, "en-us": 2, "es-419": 2, "fr-ca": 2, "uz": 2,

	// 3 plural forms
	"ru": 3, "uk": 3, "be": 3, "bs": 3, "sr": 3, "hr": 3, "sh": 3,
	"lt": 3, "lv": 3,

	// 4 plural forms
	"cs": 4, "sk": 4, "pl": 4, "sl": 4,

	// 6 plural forms
	"ar": 6,
}

// languageNameTable maps BCP 47 language tags (lowercased) to English display names.
var languageNameTable = map[string]string{
	"af":     "Afrikaans",
	"ar":     "Arabic",
	"az":     "Azerbaijani",
	"be":     "Belarusian",
	"bg":     "Bulgarian",
	"bn":     "Bengali",
	"bs":     "Bosnian",
	"ca":     "Catalan",
	"cs":     "Czech",
	"da":     "Danish",
	"de":     "German",
	"de-at":  "German (Austria)",
	"de-ch":  "German (Switzerland)",
	"el":     "Greek",
	"en":     "English",
	"en-gb":  "English (United Kingdom)",
	"en-us":  "English (United States)",
	"es":     "Spanish",
	"es-419": "Spanish (Latin America)",
	"et":     "Estonian",
	"eu":     "Basque",
	"fa":     "Persian",
	"fi":     "Finnish",
	"fr":     "French",
	"fr-ca":  "French (Canada)",
	"gl":     "Galician",
	"he":     "Hebrew",
	"hi":     "Hindi",
	"hr":     "Croatian",
	"hu":     "Hungarian",
	"hy":     "Armenian",
	"id":     "Indonesian",
	"it":     "Italian",
	"ja":     "Japanese",
	"ka":     "Georgian",
	"km":     "Khmer",
	"kn":     "Kannada",
	"ko":     "Korean",
	"lo":     "Lao",
	"lt":     "Lithuanian",
	"lv":     "Latvian",
	"ms":     "Malay",
	"my":     "Burmese",
	"nb":     "Norwegian Bokmål",
	"nl":     "Dutch",
	"nn":     "Norwegian Nynorsk",
	"pl":     "Polish",
	"pt":     "Portuguese",
	"pt-br":  "Portuguese (Brazil)",
	"ro":     "Romanian",
	"ru":     "Russian",
	"sh":     "Serbo-Croatian",
	"sk":     "Slovak",
	"sl":     "Slovenian",
	"sq":     "Albanian",
	"sr":     "Serbian",
	"sv":     "Swedish",
	"th":     "Thai",
	"tr":     "Turkish",
	"uk":     "Ukrainian",
	"ur":     "Urdu",
	"uz":     "Uzbek",
	"vi":     "Vietnamese",
	"zh":     "Chinese",
	"zh-hans": "Chinese (Simplified)",
	"zh-hant": "Chinese (Traditional)",
	"zh-tw":  "Chinese (Taiwan)",
	"zh-hk":  "Chinese (Hong Kong)",
}

// LookupNplurals returns the nplurals for the given language tag.
// Lookup is case-insensitive; tries full tag, then base language.
func LookupNplurals(tag string) (int, bool) {
	lower := strings.ToLower(tag)
	if n, ok := npluralsTable[lower]; ok {
		return n, true
	}
	// Try base language (before first '-')
	if idx := strings.IndexByte(lower, '-'); idx > 0 {
		if n, ok := npluralsTable[lower[:idx]]; ok {
			return n, true
		}
	}
	return 0, false
}

// LookupLanguageName returns the English display name for the given tag.
// Falls back to base language, then returns the raw tag if unknown.
func LookupLanguageName(tag string) string {
	lower := strings.ToLower(tag)
	if name, ok := languageNameTable[lower]; ok {
		return name
	}
	if idx := strings.IndexByte(lower, '-'); idx > 0 {
		if name, ok := languageNameTable[lower[:idx]]; ok {
			return name
		}
	}
	return tag
}

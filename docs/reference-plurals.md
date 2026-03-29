---
title: Plural Count and Language Name Reference Tables
description: >
  The embedded CLDR-derived lookup tables: BCP 47 language tag → nplurals
  count (used as fallback when Plural-Forms is absent), and language tag →
  English display name. Covers 100+ language tags including regional variants.
topics:
  - CLDR
  - nplurals
  - plural count
  - language tag
  - BCP 47
  - language display name
  - Plural-Forms fallback
  - regional variants
relevance: >
  Read when adding or correcting a language entry in plural.go, verifying
  plural count behaviour for a specific language, or debugging a "missing
  nplurals" parse error for a language you believe should be covered.
---

# Plural Count and Language Name Reference Tables

These tables are embedded in `po/plural.go`. Lookup is case-insensitive on the
full BCP 47 tag first, then falls back to the base language (e.g. `"de-AT"` →
`"de"`).

They are used only when the `Plural-Forms` header is absent or unparseable.
When `Plural-Forms` is present, its `nplurals=N` value is authoritative.

---

## nplurals Table

| Tag | nplurals | Notes |
|---|---|---|
| `af` | 2 | Afrikaans |
| `ak` | 2 | Akan |
| `am` | 2 | Amharic |
| `ar` | 6 | Arabic — unusual 6-form paradigm |
| `ar-SA` | 6 | Arabic (Saudi Arabia) |
| `az` | 2 | Azerbaijani |
| `be` | 3 | Belarusian |
| `bg` | 2 | Bulgarian |
| `bn` | 2 | Bengali |
| `br` | 5 | Breton |
| `bs` | 3 | Bosnian |
| `ca` | 2 | Catalan |
| `cs` | 3 | Czech |
| `cy` | 4 | Welsh |
| `da` | 2 | Danish |
| `de` | 2 | German |
| `de-AT` | 2 | German (Austria) |
| `de-CH` | 2 | German (Switzerland) |
| `de-DE` | 2 | German (Germany) |
| `el` | 2 | Greek |
| `en` | 2 | English |
| `en-AU` | 2 | English (Australia) |
| `en-CA` | 2 | English (Canada) |
| `en-GB` | 2 | English (United Kingdom) |
| `en-NZ` | 2 | English (New Zealand) |
| `en-US` | 2 | English (United States) |
| `en-ZA` | 2 | English (South Africa) |
| `eo` | 2 | Esperanto |
| `es` | 2 | Spanish |
| `es-AR` | 2 | Spanish (Argentina) |
| `es-CL` | 2 | Spanish (Chile) |
| `es-CO` | 2 | Spanish (Colombia) |
| `es-ES` | 2 | Spanish (Spain) |
| `es-MX` | 2 | Spanish (Mexico) |
| `et` | 2 | Estonian |
| `eu` | 2 | Basque |
| `fa` | 1 | Persian |
| `fi` | 2 | Finnish |
| `fil` | 2 | Filipino |
| `fr` | 2 | French |
| `fr-BE` | 2 | French (Belgium) |
| `fr-CA` | 2 | French (Canada) |
| `fr-CH` | 2 | French (Switzerland) |
| `fr-FR` | 2 | French (France) |
| `ga` | 5 | Irish |
| `gl` | 2 | Galician |
| `gu` | 2 | Gujarati |
| `he` | 2 | Hebrew |
| `hi` | 2 | Hindi |
| `hr` | 3 | Croatian |
| `hu` | 2 | Hungarian |
| `hy` | 2 | Armenian |
| `id` | 1 | Indonesian |
| `is` | 2 | Icelandic |
| `it` | 2 | Italian |
| `it-CH` | 2 | Italian (Switzerland) |
| `it-IT` | 2 | Italian (Italy) |
| `ja` | 1 | Japanese |
| `ka` | 1 | Georgian |
| `kk` | 2 | Kazakh |
| `km` | 1 | Khmer |
| `kn` | 2 | Kannada |
| `ko` | 1 | Korean |
| `lt` | 3 | Lithuanian |
| `lv` | 3 | Latvian |
| `mk` | 2 | Macedonian |
| `ml` | 2 | Malayalam |
| `mn` | 2 | Mongolian |
| `mr` | 2 | Marathi |
| `ms` | 1 | Malay |
| `my` | 1 | Burmese |
| `nb` | 2 | Norwegian Bokmål |
| `ne` | 2 | Nepali |
| `nl` | 2 | Dutch |
| `nl-BE` | 2 | Dutch (Belgium) |
| `nl-NL` | 2 | Dutch (Netherlands) |
| `or` | 2 | Odia |
| `pa` | 2 | Punjabi |
| `pl` | 3 | Polish |
| `pt` | 2 | Portuguese |
| `pt-BR` | 2 | Portuguese (Brazil) |
| `pt-PT` | 2 | Portuguese (Portugal) |
| `rm` | 2 | Romansh |
| `ro` | 3 | Romanian |
| `ru` | 3 | Russian |
| `si` | 2 | Sinhala |
| `sk` | 3 | Slovak |
| `sl` | 4 | Slovenian |
| `sq` | 2 | Albanian |
| `sr` | 3 | Serbian |
| `sr-Cyrl` | 3 | Serbian (Cyrillic) |
| `sr-Latn` | 3 | Serbian (Latin) |
| `sv` | 2 | Swedish |
| `sw` | 2 | Swahili |
| `ta` | 2 | Tamil |
| `te` | 2 | Telugu |
| `th` | 1 | Thai |
| `tk` | 2 | Turkmen |
| `tr` | 2 | Turkish |
| `uk` | 3 | Ukrainian |
| `ur` | 2 | Urdu |
| `uz` | 2 | Uzbek |
| `vi` | 1 | Vietnamese |
| `zh` | 1 | Chinese |
| `zh-CN` | 1 | Chinese (Simplified) |
| `zh-HK` | 1 | Chinese (Hong Kong) |
| `zh-TW` | 1 | Chinese (Traditional) |
| `zu` | 2 | Zulu |

---

## Language Display Name Table

Canonical English display names for BCP 47 tags. Used in `get` and `stats`
output as `language_name`. If a tag is not in this table, the raw tag is used
as the name.

| Tag | Display Name |
|---|---|
| `af` | Afrikaans |
| `ak` | Akan |
| `am` | Amharic |
| `ar` | Arabic |
| `ar-SA` | Arabic (Saudi Arabia) |
| `az` | Azerbaijani |
| `be` | Belarusian |
| `bg` | Bulgarian |
| `bn` | Bengali |
| `br` | Breton |
| `bs` | Bosnian |
| `ca` | Catalan |
| `cs` | Czech |
| `cy` | Welsh |
| `da` | Danish |
| `de` | German |
| `de-AT` | German (Austria) |
| `de-CH` | German (Switzerland) |
| `de-DE` | German (Germany) |
| `el` | Greek |
| `en` | English |
| `en-AU` | English (Australia) |
| `en-CA` | English (Canada) |
| `en-GB` | English (United Kingdom) |
| `en-NZ` | English (New Zealand) |
| `en-US` | English (United States) |
| `en-ZA` | English (South Africa) |
| `eo` | Esperanto |
| `es` | Spanish |
| `es-AR` | Spanish (Argentina) |
| `es-CL` | Spanish (Chile) |
| `es-CO` | Spanish (Colombia) |
| `es-ES` | Spanish (Spain) |
| `es-MX` | Spanish (Mexico) |
| `et` | Estonian |
| `eu` | Basque |
| `fa` | Persian |
| `fi` | Finnish |
| `fil` | Filipino |
| `fr` | French |
| `fr-BE` | French (Belgium) |
| `fr-CA` | French (Canada) |
| `fr-CH` | French (Switzerland) |
| `fr-FR` | French (France) |
| `ga` | Irish |
| `gl` | Galician |
| `gu` | Gujarati |
| `he` | Hebrew |
| `hi` | Hindi |
| `hr` | Croatian |
| `hu` | Hungarian |
| `hy` | Armenian |
| `id` | Indonesian |
| `is` | Icelandic |
| `it` | Italian |
| `it-CH` | Italian (Switzerland) |
| `it-IT` | Italian (Italy) |
| `ja` | Japanese |
| `ka` | Georgian |
| `kk` | Kazakh |
| `km` | Khmer |
| `kn` | Kannada |
| `ko` | Korean |
| `lt` | Lithuanian |
| `lv` | Latvian |
| `mk` | Macedonian |
| `ml` | Malayalam |
| `mn` | Mongolian |
| `mr` | Marathi |
| `ms` | Malay |
| `my` | Burmese |
| `nb` | Norwegian Bokmål |
| `ne` | Nepali |
| `nl` | Dutch |
| `nl-BE` | Dutch (Belgium) |
| `nl-NL` | Dutch (Netherlands) |
| `or` | Odia |
| `pa` | Punjabi |
| `pl` | Polish |
| `pt` | Portuguese |
| `pt-BR` | Portuguese (Brazil) |
| `pt-PT` | Portuguese (Portugal) |
| `rm` | Romansh |
| `ro` | Romanian |
| `ru` | Russian |
| `si` | Sinhala |
| `sk` | Slovak |
| `sl` | Slovenian |
| `sq` | Albanian |
| `sr` | Serbian |
| `sr-Cyrl` | Serbian (Cyrillic) |
| `sr-Latn` | Serbian (Latin) |
| `sv` | Swedish |
| `sw` | Swahili |
| `ta` | Tamil |
| `te` | Telugu |
| `th` | Thai |
| `tk` | Turkmen |
| `tr` | Turkish |
| `uk` | Ukrainian |
| `ur` | Urdu |
| `uz` | Uzbek |
| `vi` | Vietnamese |
| `zh` | Chinese |
| `zh-CN` | Chinese (Simplified) |
| `zh-HK` | Chinese (Hong Kong) |
| `zh-TW` | Chinese (Traditional) |
| `zu` | Zulu |

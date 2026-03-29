---
title: Parser Behaviour
description: >
  How the PO parser works: input requirements, parsing rules, escape sequence
  handling, plural count resolution (Plural-Forms header → CLDR → error),
  duplicate handling, and the error/warning taxonomy.
topics:
  - parser
  - PO parsing rules
  - escape sequences
  - plural count resolution
  - Plural-Forms header
  - CLDR fallback
  - nplurals
  - language tag
  - language display name
  - duplicate key handling
  - fatal errors
  - parse warnings
  - charset validation
relevance: >
  Read when modifying parser.go, debugging a parse failure or unexpected
  warning, adding support for a new PO construct, or changing how plural counts
  are resolved.
---

# Parser Behaviour

## Input

The parser accepts a byte stream (UTF-8). It must not open files itself — the
caller provides the stream.

## Parsing Rules

- Reject with a fatal error if the header's `Content-Type` charset is present
  and not `UTF-8` (case-insensitive).
- Strip `\r` from `\r\n` line endings; normalise to `\n` throughout.
- Join multi-line quoted strings: consecutive `"..."` continuation lines are
  concatenated after unescaping.
- Unescape PO escape sequences when decoding string values: `\n`, `\t`, `\r`,
  `\\`, `\"`, `\a`, `\b`, `\f`, `\v`. Unrecognised `\X` sequences pass through
  as-is with a parse warning.
- `#| msgid` and `#| msgid_plural` support continuation lines and are decoded
  identically to regular msgid values.
- The header entry is the first non-blank, non-comment entry with an empty
  msgid. Store it separately; it is not in the node list.
- Blank lines and comment-only lines between entries are consumed and not
  stored; the writer emits exactly one blank line between entries.
- Contiguous `#~` lines form a single obsolete node stored at their original
  file position.
- `msgstr[n]` indices must be contiguous starting from 0. A gap is a fatal
  error.

## Plural Count Resolution

After parsing, resolve `nplurals` in this order:

1. Parse `nplurals=N` from the `Plural-Forms` value in the header. Use if
   present and valid.
2. If absent or unparseable, look up the language tag in the embedded CLDR
   table (see [reference-plurals.md](reference-plurals.md)). Emit a parse
   warning noting the fallback.
3. If the language is also absent or unrecognised and the file contains plural
   entries → fatal error.
4. If the language is absent or unrecognised and there are no plural entries →
   set `nplurals = 1` silently.

BCP 47 tag lookup is case-insensitive. Try the full tag first (`"de-AT"`), then
the base language (`"de"`) as a fallback.

## Language Display Name

After resolving the language tag, resolve the human-readable English display
name from the embedded name table (see [reference-plurals.md](reference-plurals.md)).
Use the full tag if available, fall back to the base language. If unrecognised,
use the raw tag as the name.

## Duplicate Handling

If two non-obsolete, non-header entries share the same key, both are retained
in file order. A warning is emitted. `get` and `set` target the first match.

## Error Taxonomy

- **Fatal errors**: truncated file, illegal keyword sequence, non-UTF-8 charset
  header, non-contiguous `msgstr[n]` indices, plural entries with no resolvable
  `nplurals`. Parsing halts; partial result must not be used.
- **Warnings**: duplicate keys, unrecognised escape sequences, missing
  `Plural-Forms` with CLDR fallback. Parsing continues; warnings are included
  in command responses.

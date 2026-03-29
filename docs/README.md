---
title: pollo Documentation Index
description: >
  Index of all pollo documentation files. Each file has frontmatter
  describing its contents and when it is relevant — read that first to
  decide whether to load the full file.
---

# pollo Documentation

| File | What it covers |
|---|---|
| [overview.md](overview.md) | What pollo is, design principles, project structure, package responsibilities, non-goals |
| [po-format.md](po-format.md) | PO file constructs, entry states, counting rules, internal data model (File/Entry/Node), edge cases |
| [parser.md](parser.md) | Parser rules, escape sequences, plural count resolution, duplicate handling, error/warning taxonomy |
| [writer.md](writer.md) | Field ordering, string escaping, line wrapping, atomic writes, string identity and file-based flag convention |
| [commands.md](commands.md) | All three commands (stats/get/set), JSON schemas, flags, exit codes, placeholder validation, output format contract |
| [agent-usage.md](agent-usage.md) | LLM agent translation workflow, shell escaping recommendations, skip file format, example invocations |
| [implementation.md](implementation.md) | Go requirements, test coverage checklist, test fixture definitions (A–E), build/install instructions |
| [reference-plurals.md](reference-plurals.md) | CLDR nplurals table and language display name table for 100+ BCP 47 tags |

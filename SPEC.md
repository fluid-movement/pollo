# `pollo` — Technical Specification

> This file is a navigation stub. The specification has been split into
> focused documents in [`docs/`](docs/README.md) for easier LLM consumption.

Each document has YAML frontmatter with a `relevance` field — read that first
to decide whether the full file is needed.

| Document | Summary |
|---|---|
| [docs/overview.md](docs/overview.md) | Purpose, design principles, project structure, non-goals |
| [docs/po-format.md](docs/po-format.md) | PO constructs, entry states, data model, edge cases |
| [docs/parser.md](docs/parser.md) | Parsing rules, escapes, plural resolution, error taxonomy |
| [docs/writer.md](docs/writer.md) | Serialisation, field order, escaping, wrapping, atomic write |
| [docs/commands.md](docs/commands.md) | stats / get / set — flags, JSON schemas, exit codes, placeholder validation |
| [docs/agent-usage.md](docs/agent-usage.md) | Translation workflow, shell escaping, skip file, examples |
| [docs/implementation.md](docs/implementation.md) | Go requirements, test coverage checklist, fixture definitions, build |
| [docs/reference-plurals.md](docs/reference-plurals.md) | CLDR nplurals table, language display names |

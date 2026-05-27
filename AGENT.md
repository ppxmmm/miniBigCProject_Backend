# Codex Project Instructions

## Required Context

- Before making any project change, read every Markdown file in `docs/AI-resource/`.
- Treat those files as project-specific AI rules and apply them alongside these instructions.
- If the requested work touches database behavior, also read `database.md` before editing.
- Prefer existing project conventions over introducing new structure.

## Required Skills

- Use the `golang-pro` skill for Go implementation, refactoring, testing, concurrency, API, and service-layer work.
- Use the `database-engineer` skill for schema, migration, SQL, repository, transaction, indexing, and query-performance work.
- When a task involves both Go code and persistence behavior, use both skills together.

## Go Standards

- Follow idiomatic Go patterns already present in this repository.
- Preserve explicit error handling and wrap contextual errors where useful.
- Use `context.Context` for request-scoped or blocking operations.
- Keep interfaces small and close to the consumer.
- Run `gofmt` on changed Go files.
- Prefer table-driven tests for new or changed behavior.

## Database Standards

- Inspect existing schema, models, repositories, migrations, and query patterns before database edits.
- Prefer additive, reversible, low-risk schema changes.
- Preserve constraints, foreign keys, indexes, and transaction boundaries that protect data integrity.
- Avoid N+1 queries and unnecessary full-table scans.
- Include relevant migration or query verification steps when database behavior changes.

## Verification

- Run the narrowest relevant tests first, then broader tests when the change has wider impact.
- If verification cannot be run, state the reason and provide the exact command that should be run.

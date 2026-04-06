---
description: Synchronize all documentation sources
---

# sync-docs

## Context for AI

This command is used by opencode to keep documentation consistent after code changes in the Pakasir Go SDK. When triggered, proactively analyze changes (using git diff or tools) and update all relevant docs while preserving accuracy, Go conventions, and consistency with AGENTS.md.

## When to Run

Invoke after changes involving:
- New or modified client options (`src/client/option.go`)
- New payment methods, statuses, or constants
- Changes to service APIs (transaction, simulation, webhook)
- New or updated i18n messages
- Error handling changes (new sentinels, APIError modifications)
- New helper packages or internal types
- Public API or functional option changes
- Test or CI workflow updates

## Documentation Sources to Sync

### 1. Primary Documentation
- `README.md` (source of truth for features, usage, project structure, API coverage)
- `README.id.md` (Indonesian translation of README.md — must stay in sync)
- `AGENTS.md` (build/lint/test commands, code style for agents)
- `CONTRIBUTING.md` (contributor guide, development guidelines)
- `CONTRIBUTING.id.md` (Indonesian translation of CONTRIBUTING.md — must stay in sync)

### 2. Package-Level GoDoc
- `src/client/docs.go`
- `src/constants/docs.go`
- `src/errors/docs.go`
- `src/i18n/docs.go`
- `src/transaction/docs.go`
- `src/simulation/docs.go`
- `src/webhook/docs.go`
- `src/helper/gc/docs.go`
- `src/helper/url/docs.go`
- Any new `**/docs.go` files for new packages

### 3. i18n Translations
- `src/i18n/messages.go` — ensure every `MessageKey` has both English and Indonesian entries in the `translations` map

### 4. Code Examples
- `examples/main.go` — update if client API or usage patterns change
- Ensure README quick-start and code examples remain accurate
- Ensure godoc examples in `docs.go` files match current API

### 5. Other
- `.github/workflows/test.yml` if CI changes (OS matrix, Go version)
- `go.mod` if dependency changes affect documented requirements

When adding new services or packages, create a `docs.go`, update README "Project Structure" and "API Coverage" tables, and update AGENTS.md "Project Layout".

## Your Sync Workflow (follow exactly)

### Step 0: Analyze Changes
Use tools (grep, glob, git via bash if needed) to examine recent changes, git diff, impacted functions/options/types.

### Step 1: Update Primary Docs
Revise `README.md` and `README.id.md` (features, structure, API coverage table, payment methods table), `AGENTS.md` (commands, style guidelines, new patterns), and `CONTRIBUTING.md`/`CONTRIBUTING.id.md` (if development guidelines changed). Always update both English and Indonesian versions together to keep translations in sync.

### Step 2: Update GoDoc
Update all `docs.go` files with accurate godoc, examples, option lists matching current code. Follow existing style: `# Heading` sections, `[TypeName]` bracket references, indented code blocks.

### Step 3: Update i18n
If new user-facing errors were added, ensure `MessageKey` constants and both EN/ID translations exist in `src/i18n/messages.go`.

### Step 4: Verify & Test
- Run `go vet ./src/...`
- Run `go test -v -race ./src/...`
- Run `gofmt -s -d .` (must produce no output)
- Ensure examples in docs.go are valid and match code
- Check all public APIs are documented

### Step 5: Final Review
Output diffs/changes for review. Ensure no outdated references.

## Requirements
- Keep docs in sync with code (e.g., client options must match `option.go` and `docs.go`)
- Follow Go doc conventions (complete sentences, `[TypeName]` references)
- Maintain professional tone matching existing README/AGENTS.md
- Update AGENTS.md "Code Style" section if new patterns introduced
- Use existing patterns from neighboring `docs.go` files
- Every new package must have a `docs.go` file

## Commit Message
```
docs: sync documentation after [brief change desc]

- Update README.md, README.id.md, and AGENTS.md
- Revise package docs in docs.go files
- Verify with go vet, go test, gofmt
```

> [!NOTE]
> Use question tool if needed to confirm before committing.

## Verification Checklist
- [ ] README.md updated (features, structure, API tables)
- [ ] README.id.md updated (Indonesian translation kept in sync)
- [ ] AGENTS.md updated (commands, layout, style)
- [ ] CONTRIBUTING.md updated (if guidelines changed)
- [ ] CONTRIBUTING.id.md updated (if guidelines changed)
- [ ] `docs.go` files updated with accurate godoc
- [ ] i18n translations complete (EN + ID)
- [ ] `go vet ./src/...` passes
- [ ] `go test -v -race ./src/...` passes
- [ ] `gofmt -s -d .` produces no output
- [ ] No references to outdated features

# AGENTS.md

## Purpose

This file defines the default collaboration rules for this repository.
Until more project-specific guidance exists, contributors and agents should follow this baseline.

## Repository Baseline

- Module path: `github.com/junjieim/ledger-skill`
- Go version: `1.26.1`
- Run commands from the repository root unless a command explicitly requires another directory.
- Prefer the Go standard library first. Add third-party dependencies only when they provide clear value.
- Keep the repository clean: do not commit secrets, local environment files, build outputs, or temporary files.

## Go Command Conventions

Use native Go commands directly. No `Makefile` or task runner is defined yet.

- Format changed files with `gofmt -w <files>`
- Validate package formatting with `go fmt ./...` when packages exist
- Build all packages with `go build ./...`
- Run unit tests with `go test ./...`
- Run tests with coverage when needed using `go test -cover ./...`
- Run static checks with `go vet ./...`
- After dependency changes, run `go mod tidy`

Before opening a PR or creating a final commit, the expected minimum checks are:

- `gofmt -w` on all changed Go files
- `go test ./...`
- `go vet ./...` for non-trivial code changes
- `go mod tidy` if dependencies were added, removed, or updated

## Suggested Project Layout

Use standard Go layout unless the project later defines a different structure.

- `cmd/` for executable entrypoints
- `internal/` for private application code
- `pkg/` only for packages intended for reuse outside this repository
- `testdata/` for test fixtures

Do not create extra top-level directories without a clear reason.

## Coding Expectations

- Keep packages focused and cohesive.
- Prefer small, explicit functions over deeply coupled abstractions.
- Handle errors explicitly; do not swallow errors silently.
- Pass `context.Context` through request-scoped or I/O-bound paths when applicable.
- Prefer composition over inheritance-style patterns.
- Write comments only where intent is not obvious from the code.
- Keep public APIs stable once introduced; if behavior changes, update tests and docs in the same change.

## Testing Expectations

- New behavior should include tests unless the change is documentation-only or otherwise not testable.
- Bug fixes should include a regression test whenever practical.
- Prefer table-driven tests for logic with multiple cases.
- Keep tests deterministic and avoid network or external service dependencies unless explicitly required.

## Dependency Rules

- Keep dependencies minimal and justified.
- Prefer mature, actively maintained libraries.
- Do not add a dependency for functionality already covered well by the standard library.
- Review license and maintenance status before introducing new third-party modules.

## Git Commit Message Convention

Use Conventional Commits:

- `feat: add transaction parser`
- `fix: handle empty ledger file`
- `refactor: simplify account validation`
- `test: add coverage for balance summary`
- `docs: update setup instructions`
- `chore: initialize repository`

Rules:

- Format: `<type>(<scope>): <subject>` when a scope is useful
- Scope is optional
- Use imperative mood
- Keep the subject concise and specific
- Do not end the subject with a period
- Prefer one logical change per commit

Recommended types:

- `feat`
- `fix`
- `refactor`
- `test`
- `docs`
- `chore`
- `build`
- `ci`

## Pull Request Expectations

- Keep PRs focused and reasonably small.
- Include a short summary of what changed and why.
- Mention any follow-up work or known limitations.
- If behavior, configuration, or public interfaces change, document that in the PR description.

## Default Agent Behavior

- Do not rewrite unrelated code while implementing a focused change.
- Do not change module path, repository settings, or project layout conventions without explicit instruction.
- Prefer minimal, reviewable diffs.
- If a change requires new repo-wide rules, update this file in the same change.

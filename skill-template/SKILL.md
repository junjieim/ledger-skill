---
name: ledger-skill
description: Local CLI skill for recording, listing, searching, fetching, updating, and deleting ledger entries backed by SQLite. Use when Codex needs to manage lightweight bookkeeping data through the bundled ledger command, especially for adding a new expense, filtering entries by exact fields, searching notes, correcting a mistaken entry, or deleting a wrong record.
---

# Ledger Skill

## Overview

Use `scripts/ledger.sh` as the primary entrypoint.

Expect all data commands to print JSON on stdout. Treat `help` as the only text-only command. Let the bundled binary manage the SQLite database in the sibling `data/` directory.

## Command Selection

- Use `add` to create a new ledger entry.
- Use `list` for exact field filtering on `currency`, `category`, `from`, `to`, and `limit`.
- Use `search` for case-insensitive note matching when the user only remembers text in `note`.
- Use `get` when the entry `id` is already known.
- Use `update` to patch one or more fields on an existing entry.
- Use `delete` to remove a confirmed wrong entry.

## Working Rules

- Prefer `scripts/ledger.sh` over calling the binary path directly.
- Use `bash scripts/ledger.sh ...` when the executable bit is not preserved in the local checkout.
- Provide `datetime` as RFC3339 and `amount` as a plain decimal string.
- Normalize an empty or missing note to `none`.
- Use `list` or `search` first when the user refers to an entry indirectly and you still need its `id`.
- Use `list` for deterministic field filters and `search` for note-oriented discovery. Do not substitute one for the other.
- Read `references/cli-reference.md` when you need exact flags, field constraints, output shapes, or ordering behavior.
- Read one matching file under `examples/` when the user request is scenario-based and you want a proven command sequence.

## Response Handling

- Parse stdout JSON instead of inferring success from exit status alone.
- On success, use the `data` field as the source of truth.
- On failure, inspect `error.code` and `error.message`.
- Use `help` or `help <command>` when you need to confirm the latest local usage text.

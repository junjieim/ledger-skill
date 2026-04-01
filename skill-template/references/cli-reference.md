# Ledger CLI Reference

## Entry Model

Stored entries always contain these fields:

- `id`
- `datetime`
- `amount`
- `currency`
- `category`
- `note`
- `created_at`
- `updated_at`

## Input Rules

- Use RFC3339 for all timestamps, for example `2026-04-01T08:00:00+08:00`.
- Use a plain decimal string for `amount`, for example `10`, `10.50`, or `-5.25`.
- Do not use commas or currency symbols inside `amount`.
- Treat `currency` as a required exact-match field with allowed values: `RMB`, `HKD`, `USD`, `EUR`, `JPY`, `GBP`, `AUD`, `CAD`, `SGD`, `TWD`.
- Treat `category` as a required exact-match text field.
- If the user does not specify a currency in planning, default to `RMB`.
- Treat `note` as required at the stored-record level. Pass `none` when no real note exists.
- For `update`, provide the entry `id` plus at least one field to patch.

## Command Summary

Use these commands through `scripts/ledger.sh`:

```bash
scripts/ledger.sh add --datetime <RFC3339> --amount <decimal> --currency <currency_code> --category <text> [--note <text>]
scripts/ledger.sh list [--currency <currency_code>] [--category <text>] [--from <RFC3339>] [--to <RFC3339>] [--limit <n>]
scripts/ledger.sh search --query <text> [--limit <n>]
scripts/ledger.sh get <id>
scripts/ledger.sh update <id> [--datetime <RFC3339>] [--amount <decimal>] [--currency <currency_code>] [--category <text>] [--note <text>]
scripts/ledger.sh delete <id>
scripts/ledger.sh help [command]
```

If the local checkout does not preserve the executable bit, invoke the same commands as `bash scripts/ledger.sh ...`.

## Behavioral Notes

- `list` performs hard filtering only.
- `search` performs case-insensitive matching against `note` only.
- `list` and `search` return results ordered by `datetime DESC, created_at DESC`.
- The default SQLite path is `data/ledger.db` when the binary lives at `scripts/ledger`.

## JSON Response Shape

All data commands print a top-level object with:

- `success`
- `data`
- `error`

Success examples:

- `add`, `get`, `update`: `data` is a single entry object.
- `list`, `search`: `data` is an array of entry objects.
- `delete`: `data` is an object like `{"id":"entry-id"}`.

Failure example:

```json
{
  "success": false,
  "data": null,
  "error": {
    "code": "invalid_argument",
    "message": "amount is required"
  }
}
```

Known error codes:

- `invalid_argument`
- `not_found`
- `internal`

## Help Commands

Use these when you need the latest local usage text:

```bash
scripts/ledger.sh help
scripts/ledger.sh help add
scripts/ledger.sh help list
scripts/ledger.sh help search
scripts/ledger.sh help get
scripts/ledger.sh help update
scripts/ledger.sh help delete
```

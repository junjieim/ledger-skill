# List by Exact Fields

## Scenario

Show the most recent HKD travel entries in a fixed time window.

## Commands

```bash
scripts/ledger.sh list \
  --currency HKD \
  --category travel \
  --from 2026-04-01T00:00:00Z \
  --to 2026-04-30T23:59:59Z \
  --limit 20
```

## Use This Pattern When

Use `list` when the user gives exact filters such as currency, category, or date range and expects deterministic filtering instead of note search.

# Add a New Entry

## Scenario

Record a new lunch expense and then confirm it appears in the latest food entries.

## Commands

```bash
scripts/ledger.sh add \
  --datetime 2026-04-01T12:30:00+08:00 \
  --amount 68.00 \
  --currency HKD \
  --category food \
  --note "team lunch"

scripts/ledger.sh list --category food --limit 5
```

## Use This Pattern When

Use this sequence when the user wants to add a new record and immediately verify that it was stored.

# Search and Correct an Entry

## Scenario

The user remembers a note fragment like `taxi`, but not the entry id, and wants to correct the amount.

## Commands

```bash
scripts/ledger.sh search --query taxi --limit 10
scripts/ledger.sh update <id-from-search> --amount 85.00
scripts/ledger.sh get <id-from-search>
```

## Use This Pattern When

Use this sequence when the user identifies an entry by note text and then wants to patch one or more fields after the correct id is found.

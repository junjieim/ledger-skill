# Delete a Wrong Entry

## Scenario

The user wants to remove a duplicate or mistaken record but only knows a rough description.

## Commands

```bash
scripts/ledger.sh search --query duplicate --limit 10
scripts/ledger.sh get <id-from-search>
scripts/ledger.sh delete <id-from-search>
```

## Use This Pattern When

Use this sequence when you need to confirm the target entry before deletion and then remove it by id.

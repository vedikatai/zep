# graphiti_core — deterministic episode replay

Vendored snapshot of graphiti-core with ReplayContext support for deterministic
episode replay (UUIDs, timestamps, LLM response cache).

See `/tmp/nondeterminism-audit.md` and `/tmp/replay-report.md` on the implementer host,
and `graphiti_core/utils/replay.py`.

Run tests (requires graphiti-core deps: neo4j, pydantic, kuzu for full test, pytest-asyncio):

```
PYTHONPATH=. python3 -m pytest graphiti_core/tests/test_replay.py -v
```

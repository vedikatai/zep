# Embedding storage: float32 → float16 migration plan

## Motivation

Embedding vectors dominate memory and disk for session/document search. Switching
durable storage from IEEE-754 **float32** (4 bytes/dim) to **float16** (2 bytes/dim)
cuts vector payload size by **~50%** with acceptable recall impact for cosine search
(typical embedding values lie well inside the f16 normal range).

Compute paths (cosine similarity, MMR) still promote to **float32** at the kernel
boundary so existing SIMD (`vek32`) code keeps working.

## Inventory (this monorepo)

| Location | Role | Prior format | New format |
|---|---|---|---|
| `pkg/embeddings` | Shared codec + `Vector` type | n/a (new) | ZEP1 float16 |
| `legacy/src/models/search_common.go` | `SessionSearchResultCommon.Embedding` | `[]float32` | `embeddings.Vector` (`[]uint16` bits) |
| `legacy/src/lib/search/mmr.go` | MMR diversifier | `[]float32` in-memory | + `MaximalMarginalRelevanceF16` |
| `legacy/src/store/schema_*.go` | CE embedding tables | empty (`__embeddingTables`) | SQL migration placeholder `000000000002_*` |
| `integrations/python/zep_embeddings` | Python class for integrations | n/a | `Float16Embedding` |
| `integrations/*` agent adapters | Call Zep Cloud API only | no local vectors | no storage change (API opaque) |
| `mcp/zep-mcp-server/pkg/zep` | Cloud client wrapper | no vectors | unchanged |

There is **no** on-disk embedding corpus in CE OSS by default; migration tooling still
handles SQLite / Postgres dumps for forks and local dev data under `/tmp` and `~/.zep/data`.

## Wire / blob format (ZEP1)

```
offset  size  field
0       4     magic = "ZEP1"
4       1     version = 1
5       1     dtype  = 1 (float16) | 2 (float32 legacy)
6       4     ndims  uint32 little-endian
10      …     payload: ndims * 2 bytes (f16) or * 4 bytes (f32)
```

Raw little-endian float32 arrays **without** a header (legacy CE dumps) are also accepted
by `Decode` and rewritten as ZEP1 float16 by the migrator.

## Code changes

1. Introduce `pkg/embeddings` (`Vector`, `FromFloat32`, `Float32`, `Encode`, `Decode`, `CosineSimilarity`).
2. Change Go structs that **hold** embeddings for storage/search results to `embeddings.Vector`.
3. Keep MMR/cosine kernels on `[]float32`; add F16 entrypoints that promote once.
4. Python: `integrations/python/zep_embeddings.Float16Embedding` mirrors the codec.
5. SQL migration `000000000002_embeddings_float16` documents intent (no-op on CE).

## Offline data migration

```bash
# Scans /tmp and ~/.zep/data for *.db, *.sqlite*, *.sql, *.dump
go run ./scripts/migrations/migrate_embeddings_f16.go

# Or pass explicit paths
go run ./scripts/migrations/migrate_embeddings_f16.go /tmp/zep_embeddings.db ~/.zep/data/session.sql
```

Behaviour:

- **SQLite**: finds columns named like `embedding` / `vector` / `data`, decodes each BLOB,
  skips already-float16 and corrupt rows, updates in a transaction, continues on errors.
- **SQL / pg dumps**: rewrites `\x` hex BYTEA literals; writes `*.f32.bak` backups.
- Corrupt blobs are logged and skipped; migration does not abort the whole file set.

## Verification

```bash
# Unit + retrieval tests
cd pkg/embeddings && go test -count=1 ./...

# Vector search / MMR
cd legacy/src && go test ./lib/search/ -count=1

# Memory footprint (expect ~50% logical savings)
cd pkg/embeddings && go test -run TestMemoryFootprintSynthetic -v
go test -bench=BenchmarkMemoryFloat32VsFloat16 -benchmem
```

## Rollout

1. Land codec + type changes (this PR).
2. Run migrator on any local dumps before upgrading readers that *only* accept f16
   (current `Decode` is dual-format — safe for mixed fleets).
3. Deploy readers/writers that **emit** only float16.
4. After all nodes upgraded, optional stricter mode can reject dtype=2.

## Risks & mitigations

| Risk | Mitigation |
|---|---|
| Precision loss on extreme magnitudes | Embeddings are typically L2-normalised ~[-1,1]; f16 fine |
| Mixed-version clusters | Dual-format `Decode` during transition |
| Irreversible SQL migration | `*.f32.bak` + down migration note |
| CE has no embedding tables | Migrator no-ops; tests use synthetic DBs |

## Public actions scope

PRs and API actions target the **vedikatai** fork only (`vedikatai/zep`).

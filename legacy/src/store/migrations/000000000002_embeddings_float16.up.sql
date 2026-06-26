-- Embedding storage format migration: float32 -> float16 (ZEP1 dtype=1).
-- CE historically shipped with no embedding tables (__embeddingTables empty).
-- This migration is a no-op placeholder for deployments that later add
-- message_embeddings / document_embeddings BLOB columns; application code
-- (pkg/embeddings) reads both legacy float32 and float16 and always writes float16.
-- Offline conversion: go run ./scripts/migrations/migrate_embeddings_f16.go

SELECT 1;

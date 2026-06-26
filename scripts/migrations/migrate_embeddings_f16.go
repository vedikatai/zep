// migrate_embeddings_f16 converts embedding BLOBs from float32 (raw or ZEP1 dtype=2)
// to ZEP1 float16 (dtype=1) in SQLite databases and PostgreSQL custom/plain dumps.
//
// Usage:
//
//	go run ./scripts/migrations/migrate_embeddings_f16.go [paths...]
//
// If no paths are given, scans /tmp and $HOME/.zep/data for *.db, *.sqlite, *.sqlite3,
// *.sql, *.dump.
package main

import (
	"bytes"
	"database/sql"
	"encoding/hex"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	_ "modernc.org/sqlite"

	"github.com/getzep/zep/pkg/embeddings"
)

var (
	// Column names commonly used for embedding payloads.
	embedColRe = regexp.MustCompile(`(?i)(embedding|vector|embedding_vector|emb)`)
	// Hex-encoded BYTEA in pg dumps: \xDEADBEEF
	pgHexRe = regexp.MustCompile(`\\x([0-9a-fA-F]+)`)
)

func main() {
	paths := os.Args[1:]
	if len(paths) == 0 {
		paths = discoverDefaultPaths()
	}
	if len(paths) == 0 {
		fmt.Println("no data files found under /tmp or ~/.zep/data; nothing to migrate")
		os.Exit(0)
	}

	var migrated, skipped, failed int
	for _, p := range paths {
		fmt.Printf("==> %s\n", p)
		err := migratePath(p)
		switch {
		case err == nil:
			migrated++
			fmt.Println("    ok")
		case strings.Contains(err.Error(), "skip"):
			skipped++
			fmt.Printf("    %v\n", err)
		default:
			failed++
			fmt.Printf("    ERROR (continuing): %v\n", err)
		}
	}
	fmt.Printf("\ndone: migrated=%d skipped=%d failed=%d\n", migrated, skipped, failed)
	if migrated == 0 && failed > 0 && skipped == 0 {
		os.Exit(1)
	}
}

func discoverDefaultPaths() []string {
	roots := []string{"/tmp", filepath.Join(os.Getenv("HOME"), ".zep", "data")}
	var out []string
	for _, root := range roots {
		_ = filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
			if err != nil || d.IsDir() {
				return nil
			}
			switch strings.ToLower(filepath.Ext(path)) {
			case ".db", ".sqlite", ".sqlite3", ".sql", ".dump":
				out = append(out, path)
			}
			// also match names without extension that look like dumps
			base := strings.ToLower(filepath.Base(path))
			if strings.Contains(base, "zep") && (strings.Contains(base, "embed") || strings.Contains(base, "pg") || strings.Contains(base, "sqlite")) {
				out = append(out, path)
			}
			return nil
		})
	}
	// dedupe
	seen := map[string]struct{}{}
	var uniq []string
	for _, p := range out {
		if _, ok := seen[p]; ok {
			continue
		}
		seen[p] = struct{}{}
		uniq = append(uniq, p)
	}
	return uniq
}

func migratePath(path string) error {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".db", ".sqlite", ".sqlite3":
		return migrateSQLite(path)
	case ".sql", ".dump":
		return migrateSQLDump(path)
	default:
		// try sqlite first, then treat as dump
		if err := migrateSQLite(path); err == nil {
			return nil
		}
		return migrateSQLDump(path)
	}
}

func migrateSQLite(path string) error {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return fmt.Errorf("open sqlite: %w", err)
	}
	defer db.Close()

	rows, err := db.Query(`SELECT name FROM sqlite_master WHERE type='table' AND name NOT LIKE 'sqlite_%'`)
	if err != nil {
		return fmt.Errorf("skip: not a sqlite db (%w)", err)
	}
	var tables []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			rows.Close()
			return err
		}
		tables = append(tables, name)
	}
	rows.Close()
	if len(tables) == 0 {
		return fmt.Errorf("skip: no tables")
	}

	var converted int
	for _, table := range tables {
		cols, err := tableColumns(db, table)
		if err != nil {
			fmt.Printf("    warn table %s: %v\n", table, err)
			continue
		}
		var blobCols []string
		var pkCol string
		for _, c := range cols {
			if strings.EqualFold(c, "rowid") {
				continue
			}
			if pkCol == "" {
				pkCol = c // first column as fallback key
			}
			if embedColRe.MatchString(c) || strings.Contains(strings.ToLower(c), "blob") {
				blobCols = append(blobCols, c)
			}
		}
		// Also try any BLOB-typed column
		for _, c := range cols {
			if contains(blobCols, c) {
				continue
			}
			// probe one row type is expensive; include columns named data/payload/value
			low := strings.ToLower(c)
			if low == "data" || low == "payload" || low == "value" || low == "embedding" || low == "vector" {
				blobCols = append(blobCols, c)
			}
		}
		if len(blobCols) == 0 {
			continue
		}
		// Prefer an id-like pk
		for _, c := range cols {
			low := strings.ToLower(c)
			if low == "id" || low == "uuid" || low == "rowid" || strings.HasSuffix(low, "_id") {
				pkCol = c
				break
			}
		}
		for _, col := range blobCols {
			n, err := convertSQLiteColumn(db, table, pkCol, col)
			if err != nil {
				fmt.Printf("    warn %s.%s: %v\n", table, col, err)
				continue
			}
			converted += n
			if n > 0 {
				fmt.Printf("    %s.%s: converted %d rows\n", table, col, n)
			}
		}
	}
	if converted == 0 {
		return fmt.Errorf("skip: no embedding blobs converted")
	}
	return nil
}

func tableColumns(db *sql.DB, table string) ([]string, error) {
	rows, err := db.Query(fmt.Sprintf("PRAGMA table_info(%s)", quoteIdent(table)))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var cols []string
	for rows.Next() {
		var cid int
		var name, ctype string
		var notnull, pk int
		var dflt interface{}
		if err := rows.Scan(&cid, &name, &ctype, &notnull, &dflt, &pk); err != nil {
			return nil, err
		}
		cols = append(cols, name)
	}
	return cols, nil
}

func convertSQLiteColumn(db *sql.DB, table, pkCol, col string) (int, error) {
	q := fmt.Sprintf("SELECT %s, %s FROM %s WHERE %s IS NOT NULL",
		quoteIdent(pkCol), quoteIdent(col), quoteIdent(table), quoteIdent(col))
	rows, err := db.Query(q)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	type upd struct {
		pk  interface{}
		blob []byte
	}
	var updates []upd
	for rows.Next() {
		var pk interface{}
		var blob []byte
		if err := rows.Scan(&pk, &blob); err != nil {
			continue
		}
		if len(blob) == 0 || embeddings.IsFloat16Blob(blob) {
			continue
		}
		v, err := embeddings.Decode(blob)
		if err != nil {
			// try treating as corrupt — skip row
			fmt.Printf("    skip corrupt blob pk=%v: %v\n", pk, err)
			continue
		}
		out, err := v.Encode()
		if err != nil {
			continue
		}
		updates = append(updates, upd{pk: pk, blob: out})
	}
	tx, err := db.Begin()
	if err != nil {
		return 0, err
	}
	stmt, err := tx.Prepare(fmt.Sprintf("UPDATE %s SET %s = ? WHERE %s = ?",
		quoteIdent(table), quoteIdent(col), quoteIdent(pkCol)))
	if err != nil {
		_ = tx.Rollback()
		return 0, err
	}
	defer stmt.Close()
	for _, u := range updates {
		if _, err := stmt.Exec(u.blob, u.pk); err != nil {
			_ = tx.Rollback()
			return 0, err
		}
	}
	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return len(updates), nil
}

func migrateSQLDump(path string) error {
	raw, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	// Only rewrite \x hex blobs that decode as embeddings
	var converted int
	out := pgHexRe.ReplaceAllFunc(raw, func(m []byte) []byte {
		hexPart := m[2:] // strip \x
		blob := make([]byte, hex.DecodedLen(len(hexPart)))
		n, err := hex.Decode(blob, hexPart)
		if err != nil {
			return m
		}
		blob = blob[:n]
		if embeddings.IsFloat16Blob(blob) {
			return m
		}
		v, err := embeddings.Decode(blob)
		if err != nil {
			return m
		}
		enc, err := v.Encode()
		if err != nil {
			return m
		}
		converted++
		return append([]byte(`\x`), []byte(hex.EncodeToString(enc))...)
	})
	// Also rewrite standalone base-ish? no — keep simple
	if converted == 0 {
		// Maybe file is just a single blob written as hex lines
		if bytes.HasPrefix(bytes.TrimSpace(raw), []byte(embeddings.Magic)) || looksLikeRawFloat32(raw) {
			v, err := embeddings.Decode(raw)
			if err != nil {
				return fmt.Errorf("skip: no convertible blobs (%w)", err)
			}
			enc, err := v.Encode()
			if err != nil {
				return err
			}
			backup := path + ".f32.bak"
			if err := os.WriteFile(backup, raw, 0o644); err != nil {
				return err
			}
			if err := os.WriteFile(path, enc, 0o644); err != nil {
				return err
			}
			fmt.Printf("    wrote float16 blob (%d -> %d bytes), backup %s\n", len(raw), len(enc), backup)
			return nil
		}
		return fmt.Errorf("skip: no convertible blobs")
	}
	backup := path + ".f32.bak"
	if err := os.WriteFile(backup, raw, 0o644); err != nil {
		return err
	}
	if err := os.WriteFile(path, out, 0o644); err != nil {
		return err
	}
	fmt.Printf("    rewritten %d hex blobs, backup %s\n", converted, backup)
	return nil
}

func looksLikeRawFloat32(b []byte) bool {
	return len(b) >= 16 && len(b)%4 == 0 && !bytes.HasPrefix(b, []byte("ZEP1"))
}

func quoteIdent(s string) string {
	return `"` + strings.ReplaceAll(s, `"`, `""`) + `"`
}

func contains(ss []string, s string) bool {
	for _, x := range ss {
		if x == s {
			return true
		}
	}
	return false
}

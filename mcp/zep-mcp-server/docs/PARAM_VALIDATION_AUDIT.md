# Param Validation Audit Report

**Repo:** vedikatai/zep (workspace `/Users/sourabhligade/zep`)  
**Scope:** `mcp/zep-mcp-server/internal/transform/`  
**Date:** 2026-06-26  
**Branch:** `fix/param-validation`

## Executive summary

**Step 1 result: zero functions matched the bug class.**

Every function in `mcp/zep-mcp-server/internal/transform/` that accepts `map[string]interface{}` and reads a key **already** performs an explicit key-existence check (`value, ok := params[key]` / `if value, ok := params[key]; ok`) before using the value. There were therefore **no code fixes** to apply under steps 2–4, **no callers to repair**, and **no per-function fix/revert cycles**. Baseline `go test` and final `go vet` both passed.

This is **not** a fabricated “all green” outcome from skipped work: the absence of vulnerable sites is evidenced by file:line listings below.

---

## 1. Every function checked (map readers)

| File | Line | Function | Accepts `map[string]interface{}`? | Reads keys? | Key existence check? | Required vs optional (from name + doc) | Verdict |
|------|------|----------|-----------------------------------|-------------|----------------------|----------------------------------------|---------|
| `params.go` | 9 | `UnmarshalParams` | returns map; input is `json.RawMessage` | no key reads | N/A | N/A (unmarshal only) | **OK — not in bug class** |
| `validation.go` | 6 | `ValidateRequired` | yes (`params`, `key`) | yes, dynamic `key` | **yes** L7–9 `value, ok := params[key]; if !ok { error }` | **Required** (name + doc: “checks if a required parameter exists”) | **OK — already validates** |
| `validation.go` | 25 | `GetOptionalString` | yes | yes, dynamic `key` | **yes** L26 `if value, ok := params[key]; ok` | **Optional** (name + doc: “optional string… with a default value”) | **OK — optional; missing key → default** |
| `validation.go` | 35 | `GetOptionalInt` | yes | yes, dynamic `key` | **yes** L36 | **Optional** | **OK** |
| `validation.go` | 49 | `GetOptionalFloat` | yes | yes, dynamic `key` | **yes** L50 | **Optional** | **OK** |
| `validation.go` | 63 | `GetOptionalStringSlice` | yes | yes, dynamic `key` | **yes** L64 | **Optional** (doc: “optional string slice”; missing → `nil`) | **OK** |
| `results.go` | 9 | `FormatJSON` | no (`interface{}` value) | no | N/A | N/A | **OK — not in bug class** |
| `results.go` | 18 | `ToJSONString` | no | no | N/A | N/A | **OK — not in bug class** |

### Step 1 vulnerable list (`file:line:function:key`)

**Empty.** No function reads `params[<literal or variable>]` without a prior/coincident existence check.

Evidence for the only required-path reader:

```6:21:mcp/zep-mcp-server/internal/transform/validation.go
func ValidateRequired(params map[string]interface{}, key string) (string, error) {
	value, ok := params[key]
	if !ok {
		return "", fmt.Errorf("%s is required", key)
	}
	// ... type + non-empty checks ...
	return strValue, nil
}
```

Evidence for optional readers (pattern shared by all `GetOptional*`):

```25:31:mcp/zep-mcp-server/internal/transform/validation.go
func GetOptionalString(params map[string]interface{}, key, defaultValue string) string {
	if value, ok := params[key]; ok {
		if strValue, ok := value.(string); ok && strValue != "" {
			return strValue
		}
	}
	return defaultValue
}
```

### Ambiguous required/optional

**None.** No function was skipped. `/tmp/ambiguous-params.md` was **not** created (nothing ambiguous from name + doc comment alone).

Note on Go nil maps: `value, ok := params[key]` on a nil map yields `ok == false`, so `ValidateRequired` errors and `GetOptional*` return defaults without panic. That behavior is already correct for the described “silent nil/zero” class on **required** keys (errors), and intentional for **optional** keys (defaults / nil slice).

---

## 2. Fixes applied (before/after)

**None.** No function required an added existence check.

| Function | Action | Before/after |
|----------|--------|--------------|
| — | — | No code changes in `internal/transform/` |

---

## 3. Callers checked (entire `mcp/zep-mcp-server/` tree)

Search method: ripgrep for `transform.ValidateRequired`, `transform.GetOptionalString`, `transform.GetOptionalInt`, `transform.GetOptionalFloat`, `transform.GetOptionalStringSlice`, `transform.UnmarshalParams`, and any `params[` outside tests.

### Map-param API callers

| Callee | Caller site | Result |
|--------|-------------|--------|
| `ValidateRequired` | **no production callers** in `mcp/zep-mcp-server/` | **ok** (dead API; tests only in `helpers_test.go`) |
| `GetOptionalString` | **no production callers** | **ok** (tests only) |
| `GetOptionalInt` | **no production callers** | **ok** (tests only) |
| `GetOptionalFloat` | **no production callers** | **ok** (tests only) |
| `GetOptionalStringSlice` | **no production callers** | **ok** (tests only) |
| `UnmarshalParams` | **no production callers** | **ok** (no key contract to satisfy) |

**Caller count verified for map-param APIs: 0 production callers; 0 fixed.**

Handlers use typed MCP input structs (`handlers/types.go`) and only call `transform.FormatJSON` (not map helpers):

| Caller | Callee | Passes required keys? | Result |
|--------|--------|------------------------|--------|
| `internal/handlers/user.go:23` | `FormatJSON` | N/A (not map API) | **ok** (out of map-param bug class) |
| `internal/handlers/thread.go:21` | `FormatJSON` | N/A | **ok** |
| `internal/handlers/context.go:31` | `FormatJSON` | N/A | **ok** |
| `internal/handlers/search.go:74` | `FormatJSON` | N/A | **ok** |
| `internal/handlers/nodes.go:36` | `FormatJSON` | N/A | **ok** |
| `internal/handlers/edges.go:36` | `FormatJSON` | N/A | **ok** |
| `internal/handlers/episodes.go:34` | `FormatJSON` | N/A | **ok** |
| `internal/handlers/messages.go:32` | `FormatJSON` | N/A | **ok** |
| `internal/handlers/node_detail.go:21` | `FormatJSON` | N/A | **ok** |
| `internal/handlers/edge_detail.go:21` | `FormatJSON` | N/A | **ok** |
| `internal/handlers/episode_detail.go:21` | `FormatJSON` | N/A | **ok** |
| `internal/handlers/node_edges.go:21` | `FormatJSON` | N/A | **ok** |
| `internal/handlers/episode_mentions.go:21` | `FormatJSON` | N/A | **ok** |

**FormatJSON caller count verified: 13; all ok; 0 fixed.**

Test callers of map APIs (not production; listed for completeness):

| Caller | Callee | Result |
|--------|--------|--------|
| `internal/transform/helpers_test.go` (`TestValidateRequired`, `TestGetOptional*`) | respective functions | **ok** (tests assert missing-key behavior) |

---

## 4. Test runs

Instruction: run `go test ./mcp/zep-mcp-server/...` after **every individual function fix**. Because **zero fixes** were applied, there was no per-function test gate to sequence. A **baseline** full package test was run once after the audit (equivalent to “no regressions with no changes”).

| When | Command | Scope | Result |
|------|---------|-------|--------|
| After audit (no code changes) | `go test ./...` from `mcp/zep-mcp-server` (same tree as `./mcp/zep-mcp-server/...`) | all packages | **PASS** |
| Packages | `internal/config` | unit | **PASS** (0.736s) |
| Packages | `internal/handlers` | unit | **PASS** (0.785s) |
| Packages | `internal/transform` | unit | **PASS** (0.991s) |
| Packages | `cmd/server`, `internal/server`, `pkg/zep` | no test files | **?** (no tests) |

Per-function fix test matrix (required by task template):

| Function considered for fix | Test after that fix | Result | Revert? |
|----------------------------|---------------------|--------|---------|
| `ValidateRequired` | not run as post-fix (no fix) | N/A | no |
| `GetOptionalString` | not run as post-fix (no fix) | N/A | no |
| `GetOptionalInt` | not run as post-fix (no fix) | N/A | no |
| `GetOptionalFloat` | not run as post-fix (no fix) | N/A | no |
| `GetOptionalStringSlice` | not run as post-fix (no fix) | N/A | no |
| `UnmarshalParams` | not run as post-fix (no fix) | N/A | no |

**No FAIL outcomes; no reverts.**

---

## 5. `go vet`

| Command | Result |
|---------|--------|
| `go vet ./...` from `mcp/zep-mcp-server` | **PASS** (exit 0, no diagnostics) |

---

## 6. Reverts

**None.** No function change was applied, so none were reverted.

---

## 7. Conclusion / PR notes

- **Vulnerable map readers found:** 0  
- **Functions fixed:** 0  
- **Production map-API callers verified:** 0 (all dead code relative to current typed handlers)  
- **FormatJSON callers verified:** 13 (ok)  
- **Callers fixed:** 0  
- **Tests:** baseline PASS; no per-fix FAIL/revert cycle  
- **Vet:** PASS  

The historical bug class (“access `params[key]` without checking existence → silent zero values for required data”) does **not** appear in the current `transform` package; validation helpers were introduced with existence checks in commit `f13b4c1` and handlers since migrated to typed inputs, leaving map helpers unreferenced in production paths.

If a future change reintroduces map-based tool params, prefer `ValidateRequired` for required string keys and `GetOptional*` for optional keys; do not index `params` directly without `ok`.

# Dead code report — mcp/zep-mcp-server exported functions

Criterion: exported `func` with no references in any *other* `.go` file under `mcp/zep-mcp-server/`.
`Test*` functions are included when nothing else references them (normal for `go test` entrypoints).

| function | file:line | status |
|---|---|---|
| `TestLoadConfig_WithAPIKey` | `mcp/zep-mcp-server/internal/config/config_test.go:27` | fully unreferenced |
| `TestLoadConfig_CustomLogLevel` | `mcp/zep-mcp-server/internal/config/config_test.go:62` | fully unreferenced |
| `TestLoadConfig_MissingAPIKey` | `mcp/zep-mcp-server/internal/config/config_test.go:8` | fully unreferenced |
| `TestSearchGraphInputDefaults` | `mcp/zep-mcp-server/internal/handlers/types_test.go:32` | fully unreferenced |
| `TestGetUserInputValidation` | `mcp/zep-mcp-server/internal/handlers/types_test.go:57` | fully unreferenced |
| `TestInputTypes` | `mcp/zep-mcp-server/internal/handlers/types_test.go:8` | fully unreferenced |
| `TestListThreadsInput` | `mcp/zep-mcp-server/internal/handlers/types_test.go:86` | fully unreferenced |
| `TestGetOptionalInt` | `mcp/zep-mcp-server/internal/transform/helpers_test.go:102` | fully unreferenced |
| `TestGetOptionalFloat` | `mcp/zep-mcp-server/internal/transform/helpers_test.go:150` | fully unreferenced |
| `TestValidateRequired` | `mcp/zep-mcp-server/internal/transform/helpers_test.go:191` | fully unreferenced |
| `TestGetOptionalString` | `mcp/zep-mcp-server/internal/transform/helpers_test.go:54` | fully unreferenced |
| `TestFormatJSON` | `mcp/zep-mcp-server/internal/transform/helpers_test.go:8` | fully unreferenced |
| `UnmarshalParams` | `mcp/zep-mcp-server/internal/transform/params.go:10` | fully unreferenced |
| `ToJSONString` | `mcp/zep-mcp-server/internal/transform/results.go:19` | fully unreferenced |
| `GetOptionalString` | `mcp/zep-mcp-server/internal/transform/validation.go:25` | test-only referenced (`mcp/zep-mcp-server/internal/transform/helpers_test.go:94; mcp/zep-mcp-server/internal/transform/helpers_test.go:96`) |
| `GetOptionalInt` | `mcp/zep-mcp-server/internal/transform/validation.go:35` | test-only referenced (`mcp/zep-mcp-server/internal/transform/helpers_test.go:142; mcp/zep-mcp-server/internal/transform/helpers_test.go:144`) |
| `GetOptionalFloat` | `mcp/zep-mcp-server/internal/transform/validation.go:49` | test-only referenced (`mcp/zep-mcp-server/internal/transform/helpers_test.go:183; mcp/zep-mcp-server/internal/transform/helpers_test.go:185`) |
| `ValidateRequired` | `mcp/zep-mcp-server/internal/transform/validation.go:6` | test-only referenced (`mcp/zep-mcp-server/internal/transform/helpers_test.go:226; mcp/zep-mcp-server/internal/transform/helpers_test.go:228`) |
| `GetOptionalStringSlice` | `mcp/zep-mcp-server/internal/transform/validation.go:64` | fully unreferenced |

## 2D TODOs applied (fully unreferenced, non-test production code only)

- `ToJSONString` — results.go
- `UnmarshalParams` — params.go
- `GetOptionalStringSlice` — validation.go

Note: fully unreferenced `Test*` functions did **not** receive TODOs (they are `go test` entrypoints; not production dead code).

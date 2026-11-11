# onyx-config

[![Go Version](https://img.shields.io/badge/go-1.27-00ADD8.svg)](https://go.dev/)
[![License: MIT](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

A clean configuration-loading library for Go. Extracted from
`Onyx.Common/configmaster`.

## Installation

```bash
go get github.com/MahdiPezeshkian/onyx-config@v0.1.0
```

## Import path

```go
import "github.com/MahdiPezeshkian/onyx-config"
```

Package name: `config`

## Basic usage

```go
cfg := config.New()
cfg.SetDefault("app.port", 8080)
_ = cfg.LoadFile("config.yaml", "")
cfg.BindEnv("APP_")

port := cfg.GetInt("app.port")
host := cfg.GetString("db.host", "127.0.0.1")
```

Recommended load order: defaults → config files → `BindEnv` (highest priority).

## Public API overview

| API | Purpose |
|---|---|
| `New` | Create an empty config |
| `LoadFile` / `LoadFiles` / `LoadReader` | Load YAML, TOML, JSON, XML, or ENV |
| `Set` / `SetDefault` / `Merge` | Mutate and merge values |
| `BindEnv` | Bind prefixed environment variables (overrides existing keys) |
| `Get*` | Type-safe getters with optional fallbacks (`String`, `Int`, `Bool`, `Float`, `Duration`, `StringSlice`, `Map`) |
| `All` | Return the underlying map |

Keys use dot notation (`database.host`). `BindEnv` maps `APP_AUTH_JWT_SECRET`
onto an existing underscore key such as `auth.jwt_secret` when present.

## Design decisions

- Configuration is stored as `map[string]interface{}` (no struct unmarshalling).
- Not concurrency-safe; wrap with a mutex if shared across goroutines.
- XML parsing is intentionally flat/simple.
- Environment variables override both file values and defaults.

## Dependencies

- `gopkg.in/yaml.v3`
- `github.com/BurntSushi/toml`
- Go standard library for JSON, XML, and ENV parsing

## Limitations

- No built-in schema validation beyond typed getters.
- XML support is not a full nested document model.
- Thread safety is the caller's responsibility.

## License

MIT — see [LICENSE](LICENSE).

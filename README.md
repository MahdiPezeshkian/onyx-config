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
| `Get*` | Convenience getters with optional fallbacks |
| `LookupString` / `LookupInt` / `LookupBool` | Error-returning getters (`ErrNotFound`, `ErrTypeMismatch`) |
| `All` / `GetMap` | Deep copies of configuration maps |

Keys use dot notation (`database.host`).

### BindEnv resolution

1. Keep variables matching the prefix (e.g. `APP_`).
2. Strip the prefix and lowercase the remainder.
3. Replace `_` with `.` to form a candidate dotted key.
4. If an existing key matches by preferring literal underscore segments
   (e.g. `auth.jwt_secret` over `auth.jwt.secret`), that key is used.
5. Otherwise the plain dotted key is written.
6. Environment values always override file values and defaults.

## Design decisions

- Configuration is stored as `map[string]interface{}` (no struct unmarshalling).
- **Not concurrency-safe** for concurrent reads/writes; callers must synchronize.
- `All` and `GetMap` return deep copies so callers cannot mutate internal state.
- Existing `Get*` helpers keep their historical zero/fallback semantics; use
  `Lookup*` when conversion failures must be distinguished.
- XML parsing is a **flat** attribute/text extractor (dotted keys), not a nested
  XML document model.
- Environment variables override both file values and defaults.

## Dependencies

- `gopkg.in/yaml.v3`
- `github.com/BurntSushi/toml`
- Go standard library for JSON, XML, and ENV parsing

## Limitations

- No built-in schema validation beyond typed getters.
- XML is flat-only; nested XML maps are not produced.
- Thread safety is the caller's responsibility.
- `GetInt`/`GetBool`/… return zero on conversion failure (historical behavior).

## License

MIT — see [LICENSE](LICENSE).
